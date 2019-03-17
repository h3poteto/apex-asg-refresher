package main

import (
	"context"
	"encoding/json"
	"errors"

	log "github.com/sirupsen/logrus"
)

type TargetEvent struct {
	TargetASGs []string `json:"target_asgs"`
}

func handler(ctx context.Context, event TargetEvent) error {
	jsonBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}
	log.Info(string(jsonBytes))

	if len(event.TargetASGs) < 1 {
		return errors.New("Target asg is required")
	}

	asg := NewASG()
	err = asg.CheckGroupStatuses(event.TargetASGs)
	if err != nil {
		log.Error(err)
		return err
	}

	group, err := asg.GetASG(event.TargetASGs[0])
	if err != nil {
		return err
	}
	log.Info(*group.AutoScalingGroupName)
	err = asg.TerminateOldestInstance(group)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}
