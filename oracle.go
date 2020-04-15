package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	// "github.com/dpapathanasiou/go-recaptcha"
	"github.com/joho/godotenv"
)

type ErrorResponse struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Error   interface{} `json:"error"`
}

type SuccessResponse struct {
	Status bool        `json:"status"`
	Data   interface{} `json:"data"`
}

var chain string
var key string
var pass string
var node string
var ContractsRegistry []string

type claim_struct struct {
	Address  string `json:"address"`
	Response string `json:"response"`
}

func getEnv(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		fmt.Println(key, "=", value)
		return value
	} else {
		log.Fatal("Error loading environment variable: ", key)
		return ""
	}
}

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	chain = getEnv("CHAIN_ID")
	// recaptchaSecretKey = getEnv("FAUCET_RECAPTCHA_SECRET_KEY")
	key = getEnv("ORACLE_AC_KEY")
	pass = getEnv("ORACLE_AC_PASS")
	node = getEnv("RPC_NODE")

	initForestCoverRunner()

	// recaptcha.Init(recaptchaSecretKey)

	http.HandleFunc("/register", registerHandler)

	if err := http.ListenAndServe(getEnv("APP_URL"), nil); err != nil {
		log.Fatal("failed to start server", err)
	}
}

func executeCmd(command string, writes ...string) {
	cmd, wc, _ := goExecute(command)

	for _, write := range writes {
		wc.Write([]byte(write + "\n"))
	}
	cmd.Wait()
}

func goExecute(command string) (cmd *exec.Cmd, pipeIn io.WriteCloser, pipeOut io.ReadCloser) {
	cmd = getCmd(command)
	pipeIn, _ = cmd.StdinPipe()
	pipeOut, _ = cmd.StdoutPipe()
	go cmd.Start()
	time.Sleep(time.Second)
	return cmd, pipeIn, pipeOut
}

func getCmd(command string) *exec.Cmd {
	// split command into command and args
	split := strings.Split(command, " ")

	var cmd *exec.Cmd
	if len(split) == 1 {
		cmd = exec.Command(split[0])
	} else {
		cmd = exec.Command(split[0], split[1:]...)
	}

	return cmd
}

func registerHandler(res http.ResponseWriter, request *http.Request) {
	contractAddress := request.FormValue("contract")

	(res).Header().Set("Access-Control-Allow-Origin", "*")

	if !strings.HasPrefix(contractAddress, "xrn:") || len(contractAddress) != 43 {
		res.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(res).Encode(ErrorResponse{
			Status:  false,
			Message: "Invalid contract address",
		})
		return
	}

	ContractsRegistry = append(ContractsRegistry, contractAddress)

	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(SuccessResponse{
		Status: true,
		Data:   "Registered the contract successfully. Expect the data stream shortly.",
	})

	return
}

func initForestCoverRunner() {

	go func() {

		for {

			for _, contract := range ContractsRegistry {
				fmt.Println("Contract address:", contract)

				updateEcostateCmd := fmt.Sprintf("{\"update_ecostate\":{\"ecostate\": %d}", rand.Intn(400))

				// send the forest cover!
				sendForestCoverCmd := fmt.Sprintf(
					"xrncli tx wasm execute %v `%v` --gas auto --fee 5000utree --from %v --chain-id %v --node %v -y",
					contract, updateEcostateCmd, key, chain, node)

				fmt.Println("send command", sendForestCoverCmd)

				executeCmd(sendForestCoverCmd, pass, pass)
			}

			time.Sleep(10000 * time.Millisecond)
		}
	}()
}
