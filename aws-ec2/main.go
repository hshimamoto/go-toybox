// go-toybox/aws-ec2
// MIT License Copyright(c) 2021 Hiroshi Shimamoto
// vim:set sw=4 sts=4:
package main

import (
    "context"
    "log"

    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/ec2"
)

func main() {
    // try to load config.json
    cfg, err := config.LoadDefaultConfig(context.TODO())
    if err != nil {
	log.Printf("LoadDefaultConfig: %v\n", err)
	return
    }
    client := ec2.NewFromConfig(cfg)
    input := &ec2.DescribeInstancesInput{}
    res, err := client.DescribeInstances(context.TODO(), input)
    if err != nil {
	log.Printf("DescribeInstances: %v\n", err)
	return
    }
    // show instances
    for _, r := range res.Reservations {
	for _, i := range r.Instances {
	    id := *i.InstanceId
	    pip := "-"
	    // interfaces
	    for _, eni := range i.NetworkInterfaces {
		a := eni.Association
		if a == nil {
		    continue
		}
		if a.CarrierIp != nil {
		    pip = "carrier: " + *a.CarrierIp
		    break
		}
		if a.PublicIp != nil {
		    pip = "public: " + *a.PublicIp
		    break
		}
	    }
	    state := i.State
	    name := "-"
	    for _, tag := range i.Tags {
		if *tag.Key == "Name" {
		    name = *tag.Value
		}
	    }
	    log.Printf("%s (%s) %s %s\n", id, name, state.Name, pip)
	}
    }
}
