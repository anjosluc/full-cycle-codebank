package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/anjosluc/codebank/infrastructure/grpc/server"

	"github.com/anjosluc/codebank/infrastructure/kafka"

	"github.com/anjosluc/codebank/infrastructure/repository"
	"github.com/anjosluc/codebank/usecase"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}
}

func main() {
	db := setupDb()
	defer db.Close()

	producer := setupKafkaProducer()
	processTransactionUseCase := setupTransactionUseCase(db, producer)
	serveGrpc(processTransactionUseCase)
}

func setupKafkaProducer() kafka.KafkaProducer {
	producer := kafka.NewKafkaProducer()
	producer.SetUpProducer(os.Getenv("KafkaBootstrapServers"))
	return producer
}

func setupTransactionUseCase(db *sql.DB, producer kafka.KafkaProducer) usecase.UseCaseTransaction {
	transactionRepository := repository.NewTransactionRepositoryDb(db)
	useCase := usecase.NewUseCaseTransaction(transactionRepository)
	useCase.KafkaProducer = producer
	return useCase
}

func setupDb() *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("host"),
		os.Getenv("port"),
		os.Getenv("user"),
		os.Getenv("password"),
		os.Getenv("dbname"),
	)
	db, err := sql.Open("postgres", psqlInfo)

	if err != nil {
		log.Fatal("error connecting to db")
	}

	return db

}

func serveGrpc(processTransactionUseCase usecase.UseCaseTransaction) {
	grpcServer := server.NewGRPCServer()
	grpcServer.ProcessTransactionUseCase = processTransactionUseCase
	fmt.Println("Running grpc server")
	grpcServer.Serve()
}

/*
cc := domain.NewCreditCard()

cc.Number = 1234
cc.Name = "Lucas"
cc.CVV = 123
cc.ExpirationYear = 2022
cc.ExpirationMonth = 07
cc.Balance = 0
cc.Limit = 2000

repo := repository.NewTransactionRepositoryDb(db)
err := repo.CreateCreditCard(*cc)

if err != nil {
	fmt.Println(err)
}
*/
