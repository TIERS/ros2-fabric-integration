/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package main

//package ros2_integration
import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"

	"context"
	"sync"
	"time"

	geometry_msgs "github.com/TIERS/rclgo-msgs/geometry_msgs/msg"
	std_msgs "github.com/TIERS/rclgo-msgs/std_msgs/msg"
	"github.com/TIERS/rclgo/pkg/rclgo"
)

var pub *rclgo.Publisher
var contract *gateway.Contract

var asset_counter = 0
var last_x, last_y, last_z string

//Subscriber to the position topic
func Handler_position(s *rclgo.Subscription) {

	msg_position := geometry_msgs.PoseStampedTypeSupport.New()
	_, err := s.TakeMessage(msg_position)
	if err != nil {
		fmt.Println("failed to take message:", err)
		return
	}
	x := msg_position.(*geometry_msgs.PoseStamped).Pose.Position.X
	y := msg_position.(*geometry_msgs.PoseStamped).Pose.Position.Y
	z := msg_position.(*geometry_msgs.PoseStamped).Pose.Position.Z

	last_x = strconv.FormatFloat(x, 'f', 4, 64)
	last_y = strconv.FormatFloat(y, 'f', 4, 64)
	last_z = strconv.FormatFloat(z, 'f', 4, 64)
}

func save_to_contract(asset_id int, class_id string, probability string, last_x string, last_y string, last_z string) {

	//function for creating an asset
	count := fmt.Sprint(asset_id)
	log.Println("Object:", count, " , ", "ClassID:", class_id, " , ", "Probability:", probability, " , ", "XPosition:", last_x, " , ", "YPosition:", last_y, " , ", "ZPosition:", last_z)
	result, err := contract.SubmitTransaction("CreateObjectDetection", count, class_id, probability, last_x, last_y, last_z)
	if err != nil {
		log.Fatalf("Failed to Submit transaction: %v", err)
	}
	log.Println(string(result))

	//publisher for frequency of each asset creation process
	rosmsg := std_msgs.NewString()
	rosmsg.Data = fmt.Sprintln(class_id, ",", probability, ",", last_x, ",", last_y, ",", last_z)
	pub.Publish(rosmsg)

	//function for getting all the assets
	log.Println("--> Evaluate Transaction: GetAll Object Detections, Returns All the Current Object Detections on the Ledger")
	result, err = contract.EvaluateTransaction("GetAllObjectDetections")
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %v", err)
	}
	log.Println(string(result))
}

//subscriber to the object detection topic
func Handler_object_detection(s *rclgo.Subscription) {
	log.Println("Callback start -------> ", time.Now())
	asset_counter += 1
	msg_object_detection := std_msgs.StringTypeSupport.New()
	_, err2 := s.TakeMessage(msg_object_detection)
	if err2 != nil {
		fmt.Println("failed to take message:", err2)
		return
	}
	msg_data := msg_object_detection.(*std_msgs.String).Data
	data_list := strings.Split(msg_data, ",")
	prob := data_list[0]
	probability := prob[0:7]
	class_id := data_list[1]

	go save_to_contract(asset_counter, class_id, probability, last_x, last_y, last_z)
	log.Println("Callback done -------> ", time.Now())
}

//application starts
func main() {
	log.Println("============ application-golang starts ============")

	err := os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	if err != nil {
		log.Fatalf("Error setting DISCOVERY_AS_LOCALHOST environemnt variable: %v", err)
	}

	wallet, err := gateway.NewFileSystemWallet("wallet")
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
	}

	if !wallet.Exists("appUser") {
		err = populateWallet(wallet)
		if err != nil {
			log.Fatalf("Failed to populate wallet contents: %v", err)
		}
	}

	ccpPath := filepath.Join(
		"..",
		"..",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"connection-org1.yaml",
	)

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, "appUser"),
	)
	if err != nil {
		log.Fatalf("Failed to connect to gateway: %v", err)
	}
	defer gw.Close()

	network, err := gw.GetNetwork("mychannel")
	if err != nil {
		log.Fatalf("Failed to get network: %v", err)
	}

	contract = network.GetContract("basic")

	//Initledger Function
	log.Println("--> Submit Transaction: InitLedger, function creates the initial set of objectdetections on the ledger")
	result, err := contract.SubmitTransaction("InitLedger")
	if err != nil {
		log.Fatalf("Failed to Submit transaction: %v", err)
	}
	log.Println(string(result))

	//Get All Assets Function
	log.Println("--> Evaluate Transaction: GetAllObjectDetections, function returns all the current objectdetections on the ledger")
	result, err = contract.EvaluateTransaction("GetAllObjectDetections")
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %v", err)
	}
	log.Println(string(result))

	//Setup ROS2 nodes
	log.Println("============ ROS2 node starts ============")
	var doneChannel = make(chan bool)
	var wg sync.WaitGroup

	ctx, quitFunc := context.WithCancel(context.Background())

	rclArgs, rclErr := rclgo.NewRCLArgs("")
	if rclErr != nil {
		log.Fatal(rclErr)
	}

	rclContext, rclErr := rclgo.NewContext(&wg, 0, rclArgs)
	if rclErr != nil {
		log.Fatal(rclErr)
	}
	defer rclContext.Close()

	rclNode, rclErr := rclContext.NewNode("communicate", "publisher_test")
	if rclErr != nil {
		log.Fatal(rclErr)
	}
	//publisher to check frequency
	opts := rclgo.NewDefaultPublisherOptions()
	opts.Qos.Reliability = rclgo.RmwQosReliabilityPolicySystemDefault
	pub, err = rclNode.NewPublisher("/test_hz", std_msgs.StringTypeSupport, opts)
	if err != nil {
		log.Fatalf("Unable to create publisher: %v", err)
	}
	//subscriber to the position topic
	sub, _ := rclNode.NewSubscription("/vrpn_client_node/newdrone/pose", geometry_msgs.PoseStampedTypeSupport, Handler_position)
	go func() {
		err := sub.Spin(ctx, 1*time.Second)
		log.Printf("Subscription failed: %v", err)
	}()
	//subscriber to the Object detection topic
	sub2, _ := rclNode.NewSubscription("/object_found", std_msgs.StringTypeSupport, Handler_object_detection)
	go func() {
		err := sub2.Spin(ctx, 1*time.Second)
		log.Printf("Subscription failed: %v", err)
	}()

	log.Println("--> Evaluate Transaction: GetAllObjectDetections, function returns all the current objectdetections on the ledger")
	result, err = contract.EvaluateTransaction("GetAllObjectDetections")
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %v", err)
	}
	log.Println(string(result))

	<-doneChannel
	quitFunc()
	wg.Wait()

	log.Println("============ application-golang ends ============")
}

func populateWallet(wallet *gateway.Wallet) error {
	log.Println("============ Populating wallet ============")
	credPath := filepath.Join(
		"..",
		"..",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"users",
		"User1@org1.example.com",
		"msp",
	)

	certPath := filepath.Join(credPath, "signcerts", "cert.pem")
	// read the certificate pem
	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	keyDir := filepath.Join(credPath, "keystore")
	// there's a single file in this dir containing the private key
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		return err
	}
	if len(files) != 1 {
		return fmt.Errorf("keystore folder should have contain one file")
	}
	keyPath := filepath.Join(keyDir, files[0].Name())
	key, err := ioutil.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return err
	}

	identity := gateway.NewX509Identity("Org1MSP", string(cert), string(key))

	return wallet.Put("appUser", identity)
}
