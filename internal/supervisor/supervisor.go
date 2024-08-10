package supervisor

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/anonopiran/Fly2User/internal/config"
	"github.com/anonopiran/Fly2User/internal/v2ray"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type Supervisor struct {
	UserDB       *gorm.DB
	DownServerDB *gorm.DB
	UpSrvs       map[uint]*v2ray.UpServer
	checkUser    *v2ray.UserType
	runInterval  uint
	inbounds     *[]config.InboundConfigType
}

// ...down server manager
func (sup *Supervisor) addDownServer(upId uint, ipAddr string) error {
	downSrvDB := sup.DownServerDB
	dbRes := downSrvDB.Create(newDownServer(upId, ipAddr))
	if dbRes.Error != nil {
		return fmt.Errorf("error adding downserver to db: %s", dbRes.Error)
	}
	return nil
}
func (sup *Supervisor) rmDownServer(ipAddr string) error {
	downSrvDB := sup.DownServerDB
	dbRes := downSrvDB.Where("ip_address = ?", ipAddr).Delete(&downServer{})
	if dbRes.Error != nil {
		return fmt.Errorf("error deleting down server: %s", dbRes.Error)
	}
	return nil
}
func (sup *Supervisor) addDownServerMany(ipAddrs *mapset.Set[string], upSrvId uint, logger *logrus.Entry) {
	for dIp := range (*ipAddrs).Iter() {
		ll := logger.WithField("ip", dIp)
		ll.Info("server UP")
		if err := sup.addDownServer(upSrvId, dIp); err != nil {
			ll.WithError(err).Error("error adding new down server")
		}
	}
}
func (sup *Supervisor) rmDownServerMany(ipAddrs *mapset.Set[string], logger *logrus.Entry) {
	for dIp := range (*ipAddrs).Iter() {
		ll := logger.WithField("ip", dIp)
		ll.Info("server DOWN")
		if err := sup.rmDownServer(dIp); err != nil {
			ll.WithError(err).Error("error removing down server")
		}
	}
}
func (sup *Supervisor) GetDownServerIps(upId uint) (mapset.Set[string], error) {
	downSrvDB := sup.DownServerDB
	downSrvList := []downServer{}
	downSrvDB.Where("up_srv_id = ?", upId).Select("ip_address").Find(&downSrvList)

	ipSet := mapset.NewSet[string]()
	for _, dSrv := range downSrvList {
		ipSet.Add(dSrv.IpAddress)
	}
	return ipSet, nil
}
func (sup *Supervisor) conn(downSrv *downServer) (*grpc.ClientConn, error) {
	sup.logger(downSrv).Debug("dialing")
	return v2ray.NewInsecureGrpc(net.ParseIP(downSrv.IpAddress), sup.UpSrvs[downSrv.UpSrvId].Address.Port())
}

// ... restart manager
func (sup *Supervisor) serviceRestarted(downSrv *downServer, inbound *config.InboundConfigType, conn *grpc.ClientConn) (bool, error) {
	upSrv := sup.UpSrvs[downSrv.UpSrvId]
	ctx := context.Background()
	err := upSrv.AddUser(ctx, inbound, sup.checkUser, conn)
	if err == nil {
		return true, nil
	}
	errGrpc, ok := err.(*v2ray.GrpcError)
	if !ok || !errGrpc.IsUserExistsError() {
		return false, fmt.Errorf("error adding test user: %s", err)
	}
	return false, nil
}

// ... user manager
func (sup *Supervisor) addUser(usr *UserRecord, downSrv *downServer, ctx context.Context, conn *grpc.ClientConn) {
	upSrv := sup.UpSrvs[downSrv.UpSrvId]
	userTyped := usr.asV2ray()
	ll := sup.logger(downSrv)
	ll = usr.logger(ll)
	for _, inb := range *sup.inbounds {
		ll2 := ll.WithField("inbound", inb.Tag)
		err := upSrv.AddUser(ctx, &inb, userTyped, conn)
		if err != nil {
			if grpcErr, ok := err.(*v2ray.GrpcError); ok && grpcErr.IsUserExistsError() {
				ll2.Warn("user already exists")
			} else {
				ll2.WithError(err).Error("error adding user")
				continue
			}
		} else {
			ll2.Debug("user added")
		}
	}
}
func (sup *Supervisor) rmUser(usr *UserRecord, downSrv *downServer, ctx context.Context, conn *grpc.ClientConn) {
	upSrv := sup.UpSrvs[downSrv.UpSrvId]
	userTyped := usr.asV2ray()
	ll := sup.logger(downSrv)
	ll = usr.logger(ll)
	for _, inb := range *sup.inbounds {
		ll2 := ll.WithField("inbound", inb.Tag)
		err := upSrv.RmUser(ctx, &inb, userTyped, conn)
		if err != nil {
			if grpcErr, ok := err.(*v2ray.GrpcError); ok && grpcErr.IsUserNotFoundError() {
				ll2.Warn("user not found")
			} else {
				ll2.WithError(err).Error("error removing user")
				continue
			}
		} else {
			ll2.Debug("user removed")
		}
	}
}

// ...
func (sup *Supervisor) logger(downSrv *downServer) *logrus.Entry {
	return sup.UpSrvs[downSrv.UpSrvId].Logger(downSrv.logger(nil))
}

// ...
func (sup *Supervisor) Start() {
	sleeper := time.NewTicker(time.Second * time.Duration(sup.runInterval))
	defer sleeper.Stop()
	for {
		sup.ServiceDiscovery()
		sup.RestartHandler()
		<-sleeper.C
	}
}
func (sup *Supervisor) AddUser(usr *UserRecord) error {
	dbRes := sup.UserDB.Where(usr).FirstOrCreate(&UserRecord{})
	if dbRes.Error != nil {
		return fmt.Errorf("error adding user to db: %s", dbRes.Error)
	}
	dnSrvs := []downServer{}
	dbRes = sup.DownServerDB.Find(&dnSrvs)
	if dbRes.Error != nil {
		return fmt.Errorf("error getting downstream from DB: %s", dbRes.Error)
	}
	ctx := context.Background()
	wg := sync.WaitGroup{}
	for _, ds := range dnSrvs {
		wg.Add(1)
		ds := ds
		go func(_ds *downServer) {
			ll := sup.logger(_ds)
			ll = usr.logger(ll)
			defer wg.Done()
			conn, err := sup.conn(_ds)
			if err != nil {
				ll.WithError(err).Error("error adding user")
			}
			defer conn.Close()
			sup.addUser(usr, _ds, ctx, conn)
		}(&ds)
	}
	wg.Wait()
	return nil
}
func (sup *Supervisor) RmUser(usr *UserRecord) error {
	dbRes := sup.UserDB.Where(usr).Delete(&UserRecord{}) //no error if not exist
	if dbRes.Error != nil {
		return fmt.Errorf("error removing user from db: %s", dbRes.Error)
	}
	dnSrvs := []downServer{}
	dbRes = sup.DownServerDB.Find(&dnSrvs)
	if dbRes.Error != nil {
		return fmt.Errorf("error getting downstream from DB: %s", dbRes.Error)
	}
	ctx := context.Background()
	wg := sync.WaitGroup{}
	for _, ds := range dnSrvs {
		wg.Add(1)
		ds := ds
		go func(_ds *downServer) {
			ll := sup.logger(_ds)
			ll = usr.logger(ll)
			defer wg.Done()
			conn, err := sup.conn(_ds)
			if err != nil {
				ll.WithError(err).Error("error removing user")
			}
			defer conn.Close()
			sup.rmUser(usr, _ds, ctx, conn)
		}(&ds)
	}
	wg.Wait()
	return nil
}
func (sup *Supervisor) CountUser() (uint, error) {
	var cnt int64
	dbRes := sup.UserDB.Model(&UserRecord{}).Count(&cnt)

	if dbRes.Error != nil {
		return 0, fmt.Errorf("error getting users from DB: %s", dbRes.Error)
	}
	return uint(cnt), nil
}
func (sup *Supervisor) FlushUser() error {
	allUsers := []UserRecord{}
	dbRes := sup.UserDB.Find(&allUsers)
	if dbRes.Error != nil {
		return fmt.Errorf("error getting users from DB: %s", dbRes.Error)
	}
	dnSrvs := []downServer{}
	dbRes = sup.DownServerDB.Find(&dnSrvs)
	if dbRes.Error != nil {
		return fmt.Errorf("error getting downstream from DB: %s", dbRes.Error)
	}
	ctx := context.Background()
	wg := sync.WaitGroup{}
	for _, ds := range dnSrvs {
		wg.Add(1)
		ds := ds
		go func(_ds *downServer) {
			ll := sup.logger(_ds)
			defer wg.Done()
			conn, err := sup.conn(_ds)
			if err != nil {
				ll.WithError(err).Error("error removing user from server")
			}
			defer conn.Close()
			for _, usr := range allUsers {
				sup.rmUser(&usr, _ds, ctx, conn)
			}
		}(&ds)
	}
	wg.Wait()
	if err := sup.UserDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&UserRecord{}).Error; err != nil {
		logrus.WithError(err).Error("error flushing user db")
	}
	return nil
}
func (sup *Supervisor) ServiceDiscovery() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup
	for upSrvId, upSrv := range sup.UpSrvs {
		wg.Add(1)
		upSrvId := upSrvId
		upSrv := upSrv
		go func(_upSrv *v2ray.UpServer, _upSrvId uint) {
			defer wg.Done()
			ll := _upSrv.Logger(nil)
			newDownServers, err := _upSrv.Discover(ctx)
			ll.WithField("ips", newDownServers).Debug("discovered")
			if err != nil {
				ll.WithError(err).Error("error discovering servers")
				newDownServers = mapset.NewSet[string]()
			}
			currDownServers, err := sup.GetDownServerIps(_upSrvId)
			ll.WithField("ips", currDownServers).Debug("current")
			if err != nil {
				ll.WithError(err).Error("error getting down servers")
				return
			}
			downIp := currDownServers.Difference(newDownServers)
			upIp := newDownServers.Difference(currDownServers)
			sup.rmDownServerMany(&downIp, ll)
			sup.addDownServerMany(&upIp, _upSrvId, ll)
		}(upSrv, upSrvId)
	}
	wg.Wait()
}
func (sup *Supervisor) RestartHandler() {
	dnSrvs := []downServer{}
	if dbRes := sup.DownServerDB.Find(&dnSrvs); dbRes.Error != nil {
		logrus.WithError(dbRes.Error).Errorf("error getting downstream from DB: %s", dbRes.Error)
		return
	}
	logrus.WithField("all servers", fmt.Sprintf("%+v", dnSrvs)).Debug("Restart Handler")
	wg := sync.WaitGroup{}
	for _, ds := range dnSrvs {
		wg.Add(1)
		ds := ds
		go func(_ds *downServer) {
			defer wg.Done()
			ll := sup.logger(_ds)
			conn, err := sup.conn(_ds)
			if err != nil {
				ll.WithError(err).Error("error in restartHandler")
				return
			}
			defer conn.Close()
			rstFlag, err := sup.serviceRestarted(_ds, &((*sup.inbounds)[0]), conn)
			if err != nil {
				ll.WithError(err).Error("error checking for restart")
				return
			}
			if rstFlag {
				ll.Info("found server restart")
				allUsers := []UserRecord{}
				sup.UserDB.Find(&allUsers)
				for _, u := range allUsers {
					sup.addUser(&u, _ds, context.Background(), conn)
				}
			}
		}(&ds)
	}
	wg.Wait()
}

// ...
func NewSupervisor(cfg config.SupervisorConfigType, upstreamCfg config.UpstreamConfigType) (*Supervisor, error) {
	usrDB, err := NewUserORM(cfg.UserDB)
	if err != nil {
		return nil, fmt.Errorf("error creating supervisor (user db): %s", err)
	}
	downServerdb, err := newDownServerDB()
	if err != nil {
		return nil, fmt.Errorf("error creating supervisor (down server db): %s", err)
	}
	upServList := map[uint]*v2ray.UpServer{}
	for c, u := range upstreamCfg.Address {
		upServList[uint(c)], err = v2ray.NewServer(u)
		if err != nil {
			return nil, fmt.Errorf("can not create upstream server: %s", err)
		}
	}
	sup := Supervisor{
		runInterval:  cfg.Interval,
		UserDB:       usrDB,
		UpSrvs:       upServList,
		DownServerDB: downServerdb,
		checkUser: &v2ray.UserType{
			Email:  "supervisor@v2ray.com",
			Secret: uuid.NewString(),
			Level:  0,
		},
		inbounds: &upstreamCfg.InboundList,
	}
	return &sup, nil
}
