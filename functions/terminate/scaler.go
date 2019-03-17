package main

import (
	"errors"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	log "github.com/sirupsen/logrus"
)

type ASG struct {
	awsASG autoscalingiface.AutoScalingAPI
	awsEC2 ec2iface.EC2API
}

func NewASG() *ASG {
	awsASG := autoscaling.New(session.New())
	awsEC2 := ec2.New(session.New())
	return &ASG{
		awsASG,
		awsEC2,
	}
}

func (a *ASG) GetASG(name string) (*autoscaling.Group, error) {
	input := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{
			aws.String(name),
		},
	}
	output, err := a.awsASG.DescribeAutoScalingGroups(input)
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

func (a *ASG) CheckGroupStatuses(names []string) error {
	for _, name := range names {
		group, err := a.GetASG(name)
		if err != nil {
			return err
		}
		err = a.ConfirmGroupStatus(group)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *ASG) ConfirmGroupStatus(target *autoscaling.Group) error {
	// TODO: check instance state
	if len(target.Instances) == 0 {
		return fmt.Errorf("There are no instances in %s", *target.AutoScalingGroupName)
	}
	if len(target.Instances) != int(*target.DesiredCapacity) {
		return fmt.Errorf("Instances are insufficient in %s", *target.AutoScalingGroupName)
	}
	return nil
}

func (a *ASG) TerminateOldestInstance(target *autoscaling.Group) error {
	instance, err := a.getOldestInstance(target)
	if err != nil {
		return err
	}
	log.Infof("Terminationg an instance: %s", *instance.InstanceId)
	input := &ec2.TerminateInstancesInput{
		InstanceIds: []*string{
			instance.InstanceId,
		},
	}
	_, err = a.awsEC2.TerminateInstances(input)
	if err != nil {
		return err
	}
	return nil
}

func (a *ASG) getOldestInstance(target *autoscaling.Group) (*ec2.Instance, error) {
	instanceIds := []*string{}
	for _, instance := range target.Instances {
		instanceIds = append(instanceIds, instance.InstanceId)
	}
	input := &ec2.DescribeInstancesInput{
		InstanceIds: instanceIds,
	}
	output, err := a.awsEC2.DescribeInstances(input)
	if err != nil {
		return nil, err
	}
	instances := []*ec2.Instance{}
	for _, reservation := range output.Reservations {
		for _, instance := range reservation.Instances {
			instances = append(instances, instance)
		}
	}
	sort.Slice(instances, func(i, j int) bool {
		t := *instances[i].LaunchTime
		return t.Before(*instances[j].LaunchTime)
	})
	return instances[0], nil
}
