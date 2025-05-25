package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	"gitlab.com/pisya-dev/account-service/pkg/api/account_service"
	"gitlab.com/pisya-dev/promo-code-service/internal/config"
	promogrpc "gitlab.com/pisya-dev/promo-code-service/internal/grpc"
	accountserviceclient "gitlab.com/pisya-dev/promo-code-service/internal/grpc/client/account_service"
	promoHandler "gitlab.com/pisya-dev/promo-code-service/internal/grpc/handler/promo"
	"gitlab.com/pisya-dev/promo-code-service/internal/grpc/interceptor"
	promoService "gitlab.com/pisya-dev/promo-code-service/internal/service/promo"
	"gitlab.com/pisya-dev/promo-code-service/internal/storage/promo"
	"gitlab.com/pisya-dev/promo-code-service/internal/storage/promo_code"
	promopb "gitlab.com/pisya-dev/promo-code-service/pkg/api/pb"
	"gitlab.com/pisya-dev/promo-code-service/pkg/migrations"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

func main() {

	const op = "cmd.Main"

	cfg := config.MustLoad()

	fmt.Println(cfg)
	log, err := setupLogger(envLocal)
	if err != nil {
		panic(fmt.Errorf("%s: failed to setup logger: %s", op, err))
	}

	log.Info("Starting Promo Service", zap.Int("port", cfg.GRPCPort))

	redisDb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.RedisHost, cfg.RedisPort), // или "redis:6379" в docker-compose
		Password: "",                                                 // если нет пароля
		DB:       0,                                                  // используем 0-ю базу
	})
	err = migrations.Migrate(cfg)
	if err != nil {
		log.Fatal("failed mgrations", zap.Error(err))
	}

	db, err := sqlx.Open(
		"pgx",
		fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s",
			cfg.PostgresUser,
			cfg.PostgresPassword,
			cfg.PostgresHost,
			cfg.PostgresPort,
			cfg.PostgresDb,
		),
	)

	if err != nil {
		panic(fmt.Errorf("%s: failed to connect to database:\n%s", op, err))
	}

	promoRepository := promo.New(db)

	promoCodeRepository := promo_code.New(db)

	accountServiceGRPCConnect, err := grpc.NewClient(cfg.AccountServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(fmt.Errorf("grpc.NewClient: failed to create account service client: %s", err))
	}

	accountServiceGRPCClient := account_service.NewAccount_ServiceClient(accountServiceGRPCConnect) //account_service.NewAccount_ServiceClient(accountServiceGRPCConnect)

	accountServiceClient := accountserviceclient.NewClient(accountServiceGRPCClient)

	promoS := promoService.New(log, promoRepository, promoCodeRepository, redisDb, accountServiceClient)

	promoH := promoHandler.New(promoS)

	server := grpc.NewServer(grpc.UnaryInterceptor(interceptor.AuthInterceptor))

	serverAPI := promogrpc.New(promoH)

	promopb.RegisterPromoServiceServer(server, serverAPI)

	go mustRunGRPCServer(server, cfg.GRPCPort)

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop

	log.Info("Stopping application...\n", zap.String("signal", sign.String()))

	server.GracefulStop()

	log.Info("Application stopped")
}

func setupLogger(env string) (*zap.Logger, error) {

	var log *zap.Logger
	var err error

	switch env {
	case envProd:
		log, err = zap.NewProduction()
	case envLocal:
		log, err = zap.NewDevelopment()
	default:
		return nil, fmt.Errorf("unknown environment %q", env)
	}
	return log, err
}

func mustRunGRPCServer(server *grpc.Server, port int) {

	const op = "main.RunGRPCServer"

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(fmt.Errorf("%s: failed to listen: %s", op, err))
	}

	err = server.Serve(lis)
	if err != nil {
		panic(fmt.Errorf("%s: failed to serve: %s", op, err))
	}

}
