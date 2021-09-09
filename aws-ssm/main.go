// go-toybox/aws-ssm
// MIT License Copyright(c) 2021 Hiroshi Shimamoto
// vim:set sw=4 sts=4:
package main

import (
    "context"
    "log"
    "os"
    "time"

    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/ssm"
)

func main() {
    if len(os.Args) != 2 {
	return
    }
    // try to load config.json
    cfg, err := config.LoadDefaultConfig(context.TODO())
    if err != nil {
	log.Printf("LoadDefaultConfig: %v\n", err)
	return
    }
    client := ssm.NewFromConfig(cfg)
    target := os.Args[1]
    doc := "AWS-StartSSHSession"
    params := map[string][]string{}
    params["portNumber"] = []string{"22"}
    input := &ssm.StartSessionInput{
	Target: &target,
	DocumentName: &doc,
	Parameters: params,
    }
    res, err := client.StartSession(context.TODO(), input)
    if err != nil {
	log.Printf("StartSession: %v\n", err)
	return
    }
    // show output
    str := func(p *string) string {
	if p != nil {
	    return *p
	}
	return ""
    }
    log.Printf(`SessionId: %s
StreamUrl: %s
TokenValue: %s
`, str(res.SessionId), str(res.StreamUrl), str(res.TokenValue))

    log.Println("Use session-manager-plugin with the above resp")
    time.Sleep(time.Minute)
}
