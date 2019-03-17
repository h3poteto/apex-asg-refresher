package main

import (
	"context"
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

type TargetEvent struct {
	TargetASG string `json:"target_asg"`
}

func handler(ctx context.Context, event TargetEvent) error {
	jsonBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}
	log.Info(string(jsonBytes))

	asg := NewASG()
	group, err := asg.GetASG(event.TargetASG)
	if err != nil {
		return err
	}
	log.Info(*group.AutoScalingGroupName)
	return nil
}
