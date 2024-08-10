package config

type SupervisorConfigType struct {
	Interval uint   `koanf:"interval" validate:"required"`
	UserDB   string `koanf:"user_db" validate:"filepath"`
}
