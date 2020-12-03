package logger

import (
	"fmt"
	"net"
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
)

// Logger is a global object that write logs
var Logger zerolog.Logger

func init() {
	godotenv.Load()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMicro

	// logFile, err := os.OpenFile(os.Getenv("LOGGER_FILE"), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	// if err != nil {
	// 	panic(err)
	// }

	con, err := net.Dial("udp", os.Getenv("LOGGER_URL"))
	if err != nil {
		fmt.Print("errror")
	}
	// consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}

	multi := zerolog.MultiLevelWriter(con)
	Logger = zerolog.New(multi).With().
		Array("tags", zerolog.Arr().Str(os.Getenv("MS_NAME"))).
		Timestamp().Logger()
}
