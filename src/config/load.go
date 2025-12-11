package config

const (
	TurnifyUrlEnv               = "TURNIFY_URL"
	TurnifyPortEnv              = "TURNIFY_PORT"
	MatrixHomeserverUrlEnv      = "TURNIFY_MATRIX_HOMESERVER_URL"
	CFTurnTokenIDEnv            = "TURNIFY_CF_TURN_TOKEN_ID"
	CFTurnApiTokenEnv           = "TURNIFY_CF_TURN_API_TOKEN"
	TurnCredentialTTLSecondsEnv = "TURNIFY_TURN_CREDENTIAL_TTL_SECONDS"
)

type Config struct {
	TurnifyUrl               string
	TurnifyPort              int
	MatrixHomeserverUrl      string
	CFTurnTokenID            string
	CFTurnApiToken           string
	TurnCredentialTTLSeconds int
}

func New() Config {
	return Config{
		getEnv(TurnifyUrlEnv, "http://localhost:4499"),
		getEnvInt(TurnifyPortEnv, 4499),
		getEnv(MatrixHomeserverUrlEnv, "http://tuwunel:8008"),
		getEnv(CFTurnTokenIDEnv),
		getEnv(CFTurnApiTokenEnv),
		getEnvInt(TurnCredentialTTLSecondsEnv, 86400),
	}
}
