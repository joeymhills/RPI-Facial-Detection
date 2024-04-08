package client

/*
#include <wiringPi.h>
#include <unistd.h>
#include <stdlib.h>
#include <stdio.h>

int motionSensor() {
  // Initialize WiringPi library
  if (wiringPiSetup() == -1) {
    fprintf(stderr, "Failed to initialize WiringPi\n");
  return 1;
}
//Sets pin 22 to input mode
pinMode(22, INPUT)

//waits for motion to be detected
while(digitalRead(22)){
  sleep(.1);
return 0;
}

}
*/
import "C"
import (
  "log"
  "os"
  "time"
  "os/exec"
  "context"

  "google.golang.org/grpc"
  "google.golang.org/grpc/credentials/alts"
  pb "github.com/joeymhills/rpi-facial-detection/proto"
)

type imageClient struct{
  pb.UnimplementedImageServiceServer
}

func sendImage() {

  //Executes libcamera to capture image 
  cmd := exec.Command("libcamera-still -o img/temp.jpg")
  err := cmd.Run()
  if err != nil {
    log.Println("error with cli:", err)
    return
  }

  //address for google vm
  addr := "34.68.52.223:443"
  imagePath := "img/temp.jpg"

  //Reads data from imagePath
  imageData, err := os.ReadFile(imagePath)
  if err != nil {
    log.Println("error reading image data:", err) 
  }

  // Set up a connection to the server
  conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(alts.NewClientCreds(alts.DefaultClientOptions())))
  if err != nil {
    log.Fatalln("Failed to dial server:", err)
  }
  defer conn.Close()

  //Creates client gRPC client
  client := pb.NewImageServiceClient(conn)

  ctx := context.Background()
  req := &pb.ImageRequest{
    ImageData: imageData,
  }
  resp, err := client.UploadImage(ctx, req)
  if err != nil {
    log.Fatalln("error in sending image to server:", err)
  }

  log.Println("Response from server:", resp)
}

func WaitForMotion() {
  //Calls C code that waits for motion
  if C.motionSensor() == 0 {
    //Once motion is sensed we take a picture
    log.Println("motion sensed")
    sendImage()

    //Sleep to prevent taking too many pictures
    time.Sleep(time.Second*2)
    
    //Recursively call WaitForMotion() to reinstate motion detection mode
    WaitForMotion()
  }
}
