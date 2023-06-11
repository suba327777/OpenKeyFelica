package main

import (
	"fmt"
	"github.com/bamchoh/pasori"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stianeikeland/go-rpio/v4"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	managePWMPin rpio.Pin
	isOpenKey    bool
	lock         bool   = false
	isRegister   bool   = false
	tempName     string = ""
)

const (
	DebugLogPrefix        = "[DEBUG]"
	PwmPin                = 13
	VID            uint16 = 0x054C // SONY
	PID            uint16 = 0x06C1 // RC-S380
	Debug                 = true
)

func main() {
	log.Printf("%s /////// START OPEN KEY PROCESS ///////\n", DebugLogPrefix)

	initialize()

	initializeRestApiServer()

	for {
		// sudoしないと動かないので注意
		idm, err := pasori.GetID(VID, PID)
		if err != nil {
			continue
		}

		log.Println(idm)

		if isRegister {
			log.Println("Start Register User")
			id, _ := uuid.NewUUID()
			user := User{
				ID:          id.String(),
				IDM:         idm,
				Name:        tempName,
				LastLogging: "",
				StNum:       "",
			}
			SaveUserData(user)
			userData = append(userData, user)
			log.Println("End Register User")
			isRegister = false
			time.Sleep(1000 * time.Millisecond)
			continue
		}

		if Contains(userData, idm) {
			if isOpenKey {
				CloseKey()
			} else {
				OpenKey()
			}
		}
		if isOpenKey {
			CloseKey()
		} else {
			OpenKey()
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

func initialize() {
	log.Printf("%s -: Initializing -----\n", DebugLogPrefix)

	////////////////// SERVO

	fmt.Println("-: -: Servo setup...")
	err := rpio.Open()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	managePWMPin = rpio.Pin(PwmPin) // SEIGYO OUT PUT PIN
	managePWMPin.Mode(rpio.Pwm)
	managePWMPin.Freq(50 * 100)
	managePWMPin.DutyCycle(0, 100)
	fmt.Println("-: -: END Servo setup")

	////////////////// PASORI

	fmt.Println("-: -: IDM Read setup...")

	// 登録されているIDM読み取り処理
	ReadUserData()

	fmt.Println("-: -: END IDM Read setup")

}

func OpenKey() {
	lock = true
	for i := 1; i <= 60; i++ {
		managePWMPin.DutyCycle(uint32(i), 100)
		time.Sleep(10 * time.Millisecond)
	}
	isOpenKey = true
}

func CloseKey() {
	for i := 1; i <= 60; i++ {
		managePWMPin.DutyCycle(uint32(50-i), 100)
		time.Sleep(10 * time.Millisecond)
	}
	isOpenKey = false
}

func initializeRestApiServer() {
	router := gin.Default()
	router.GET("/user", getUser)
	router.POST("/user", postUser)
	go router.Run("localhost:8080")
	//go router.Run("localhost:8080")
}

func getUser(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, userData)
}

func postUser(c *gin.Context) {
	isRegister = true
	tempName = c.Params.ByName("name")
	c.IndentedJSON(http.StatusOK, tempName)
}
