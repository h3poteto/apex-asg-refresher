package main

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
)

type ASG struct {
	client autoscalingiface.AutoScalingAPI
}

func NewASG() *ASG {
	client := autoscaling.New(session.New())
	return &ASG{
		client,
	}
}

func (a *ASG) GetASG(name string) (*autoscaling.Group, error) {
	input := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{
			aws.String(name),
		},
	}
	output, err := a.client.DescribeAutoScalingGroups(input)
	if err != nil {
		return nil, err
	}
	if len(output.AutoScalingGroups) == 0 {
		return nil, errors.New("Specified group is not found")
	}
	if len(output.AutoScalingGroups) > 1 {
		return nil, errors.New("Too many groups are found")
	}
	return output.AutoScalingGroups[0], nil
}

func (a *ASG) ScaleUp(target *autoscaling.Group, numberOfIncrease int64) error {
	capacity := *target.DesiredCapacity + numberOfIncrease
	input := &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: target.AutoScalingGroupName,
		DesiredCapacity:      aws.Int64(capacity),
	}
	_, err := a.client.UpdateAutoScalingGroup(input)
	return err
}

func (a *ASG) ScaleDown(target *autoscaling.Group) error {
	return nil
}
