package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/sns"
)

func getTime() string {
	// Get current time
	currentTime := time.Now()
	t := currentTime.Format("2006-01-02 15:04:05")
	// Replace space with underscore
	t = strings.Replace(t, " ", "-", -1)
	t = strings.Replace(t, ":", "-", -1)
	return t
}

func isSnapshotTagged(tagKey string, tags []*rds.Tag) bool {
	for _, tag := range tags {
		if aws.StringValue(tag.Key) == tagKey {
			return true
		}
	}
	return false
}

func createDBClusterSnapshot(svc *rds.RDS, cluster string, snapshotIdentifier string, targetAccId []string) string {
	// Create the RDS Cluster snapshot

	snapshot, err := svc.CreateDBClusterSnapshot(&rds.CreateDBClusterSnapshotInput{
		DBClusterIdentifier:         aws.String(cluster),
		DBClusterSnapshotIdentifier: aws.String(snapshotIdentifier),
	})
	if err != nil {
		exitErrorf("Unable to create snapshot in cluster %q, %v", cluster, err)
	}

	fmt.Printf("Creating snapshot %s\n", *snapshot.DBClusterSnapshot.DBClusterSnapshotArn)

	// Wait until snapshot is created before finishing
	fmt.Printf("Waiting for snapshot in cluster %q to be created...\n", cluster)

	err = svc.WaitUntilDBClusterSnapshotAvailable(&rds.DescribeDBClusterSnapshotsInput{
		DBClusterIdentifier:         aws.String(cluster),
		DBClusterSnapshotIdentifier: aws.String(snapshotIdentifier),
	})
	if err != nil {
		exitErrorf("Error occurred while waiting for snapshot to be created in cluster, %v", cluster)
	}

	fmt.Printf("Snapshot %q successfully created in cluster\n", cluster)
	fmt.Printf("Tagging snapshot\n")

	_, err = svc.AddTagsToResource((&rds.AddTagsToResourceInput{
		ResourceName: snapshot.DBClusterSnapshot.DBClusterSnapshotArn,
	}).SetTags(
		[]*rds.Tag{
			&rds.Tag{
				Key:   aws.String("lambda_automatic"),
				Value: aws.String("true"),
			},
		},
	))
	if err != nil {
		exitErrorf("Error tagging snapshot, %v", err)
	}

	_, err = svc.ModifyDBClusterSnapshotAttribute(&rds.ModifyDBClusterSnapshotAttributeInput{
		AttributeName:               aws.String("restore"),
		DBClusterSnapshotIdentifier: aws.String(snapshotIdentifier),
		ValuesToAdd:                 aws.StringSlice(targetAccId),
	})
	if err != nil {
		exitErrorf("Failed to share snapshot with another account, %v", cluster)
	}

	fmt.Printf("Snapshot %q successfully shared with target account\n", cluster)

	return *snapshot.DBClusterSnapshot.DBClusterSnapshotArn

}

func publishMessage(svc *sns.SNS, message string, arn string, snapshot_identifier string, snapshotArn string) {

	fmt.Printf("Publishing message to queue")

	result, err := svc.Publish(&sns.PublishInput{
		Message:  aws.String(message),
		TopicArn: aws.String(arn),
		MessageAttributes: map[string]*sns.MessageAttributeValue{
			"snapshot_identifier": &sns.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String(snapshot_identifier),
			},
			"snapshot_arn": &sns.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String(snapshotArn),
			},
		},
	})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println(*result.MessageId)
}

type Request struct {
}

func removeOldSnapshots(svc *rds.RDS, clusterId string, retentionDays int) {
	dbInput := (&rds.DescribeDBClusterSnapshotsInput{
		DBClusterIdentifier: aws.String(clusterId),
	}).SetFilters([]*rds.Filter{
		&rds.Filter{
			Name:   aws.String("snapshot-type"),
			Values: []*string{aws.String("manual")},
		},
	})

	result, err := svc.DescribeDBClusterSnapshots(dbInput)
	if err != nil {
		exitErrorf("Unable to list snapshots, %v", err)
	}

	currentTime := time.Now()

	for _, s := range result.DBClusterSnapshots {
		timeDiff := currentTime.Sub(*s.SnapshotCreateTime)

		result, err := svc.ListTagsForResource(&rds.ListTagsForResourceInput{
			ResourceName: s.DBClusterSnapshotArn,
		})
		if err != nil {
			exitErrorf("Unable to get tags for snapshot, %v", err)
		}

		if int(timeDiff.Hours()/24) >= retentionDays && isSnapshotTagged("lambda_automatic", result.TagList) {
			fmt.Printf("Deleting snapshot %s from %s\n",
				aws.StringValue(s.DBClusterSnapshotIdentifier), s.SnapshotCreateTime.Format("2006-01-02 15:04:05"))

			_, err := svc.DeleteDBClusterSnapshot(&rds.DeleteDBClusterSnapshotInput{
				DBClusterSnapshotIdentifier: s.DBClusterSnapshotIdentifier,
			})
			if err != nil {
				exitErrorf("Unable to delete snapshot, %v", err)
			}
		}
	}
}

func HandleRequest(ctx context.Context, req Request) (string, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("REGION"))},
	)
	if err != nil {
		log.Fatal("Error creating session")
	}

	// Create RDS service client
	svc := rds.New(sess)
	snapshotIdentifier := os.Getenv("CLUSTER_ID") + getTime()
	targetAccounts := []string{os.Getenv("TARGET_ACCOUNT_ID")}
	snsTopic := os.Getenv("SNS_TOPIC_ARN")

	snapshotArn := createDBClusterSnapshot(svc, os.Getenv("CLUSTER_ID"), snapshotIdentifier, targetAccounts)

	svc_sns := sns.New(sess)
	publishMessage(svc_sns, "Copy snapshot"+snapshotArn, snsTopic, snapshotIdentifier, snapshotArn)

	retentionDays, err := strconv.Atoi(os.Getenv("RETENTION_DAYS"))
	if err != nil {
		log.Fatal("Error parsing RETENTION_DAYS env var")
	}

	removeOldSnapshots(svc, os.Getenv("CLUSTER_ID"), retentionDays)
	return "Finished", nil
}

func main() {
	lambda.Start(HandleRequest)
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
