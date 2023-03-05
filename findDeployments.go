package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/xxthegonzxx/s3/actions"
)

var (
	awsRegion       string
	awsEndpoint     string
	bucketName      string
	s3Client        *s3.Client
	objectsToDelete []string
)

type Object struct {
	Name         string
	LastModified time.Time
}

// ByDate implements sort.Interface for []Object based on
// the LastModified field.
type ByDate []Object

func (o ByDate) Len() int           { return len(o) }
func (o ByDate) Swap(i, j int)      { o[i], o[j] = o[j], o[i] }
func (o ByDate) Less(i, j int) bool { return o[i].LastModified.Before(o[j].LastModified) }

func init() {
	awsRegion = os.Getenv("AWS_REGION")
	awsEndpoint = os.Getenv("AWS_ENDPOINT")
	bucketName = os.Getenv("S3_BUCKET")

	// Defaults
	awsRegion = "us-east-1"
	awsEndpoint = "http://localhost:4566"
	bucketName = "sample-bucket"

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if awsEndpoint != "" {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           awsEndpoint,
				SigningRegion: awsRegion,
			}, nil
		}

		// returning EndpointNotFoundError will allow the service to fallback to it's default resolution
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(awsRegion),
		config.WithEndpointResolverWithOptions(customResolver),
	)

	if err != nil {
		log.Fatalf("Cannot load the AWS configs: %s", err)
	}

	s3Client = s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})
}

// CreateTestObjects creates test S3 objects for testing.
func CreateTestObjects(s3action actions.BucketActions, prefix, suffix []string) {
	for i := 0; i < len(prefix); i++ {
		for j := 0; j < len(suffix); j++ {
			s3action.CreateObjects(bucketName, prefix[i]+suffix[j])
		}
	}
}

// CreateHashMap lists and sorts objects based on LastModified Date.
// It takes x as a parameter and returns the most recent x deployments.
func CreateHashMap(objects []types.Object) map[string]time.Time {
	keyMap := make(map[string]time.Time)
	// creates a hash map of prefixes with most recent modified date
	for _, object := range objects {
		var prefix = strings.SplitN(*object.Key, "/", 2)[0]
		currentValue, ok := keyMap[prefix]
		//checks whether the date is in the hash map and overwrites it if its modified date is the most recent
		if !ok {
			keyMap[prefix] = *object.LastModified
		} else {
			switch currentValue.Compare(*object.LastModified) {
			case -1:
				keyMap[prefix] = *object.LastModified
			default:
				continue
			}
		}
	}
	return keyMap
}

// SortHashMap returns a sorted HashMap
func SortHashMap(keyMap map[string]time.Time) ByDate {
	// sorts hash map by modified date
	keyMapSorted := make(ByDate, len(keyMap))
	i := 0
	for k, v := range keyMap {
		keyMapSorted[i] = Object{k, v}
		i++
	}
	sort.Sort(keyMapSorted)
	return keyMapSorted
}

// FindMostRecentDeployments takes in a parameter x and returns
// a hashmap of the most recent deployments.
func FindMostRecentDeployments(keyMapSorted ByDate, x int) map[string]time.Time {
	if x == 0 {
		log.Fatalf("No Deployments selected.\nExiting...")
	}
	if x < keyMapSorted.Len() {
		keyMapSorted = keyMapSorted[:x]
	}
	mostRecentDeployments := make(map[string]time.Time)
	// converts sorted list back to hash map
	fmt.Println("Most Recent Deployments:")
	for _, k := range keyMapSorted {
		fmt.Printf("%v\t%v\n", k.Name, k.LastModified)
		mostRecentDeployments[k.Name] = k.LastModified
	}
	return mostRecentDeployments
}

// PopulateObjectsToDelete populates a list of objects to delete based on
// whether or not their prefix is in the mostRecentDeployments hash map.
func PopulateObjectsToDelete(objects []types.Object, recentDeployments map[string]time.Time) []string {
	for _, object := range objects {
		var prefix = strings.SplitN(*object.Key, "/", 2)[0]
		_, ok := recentDeployments[prefix]
		if ok {
			continue
		} else {
			objectsToDelete = append(objectsToDelete, *object.Key)
		}
	}
	return objectsToDelete
}

func main() {
	bucketActions := actions.BucketActions{S3Client: s3Client}
	// Adds flags to program
	numbPtr := flag.Int("deploys", 0, "Number of most recent deployments to keep.")
	toDelete := flag.Bool("delete", false, "Deletes the S3 objects.")
	flag.Parse()
	// Uncomment to create sample test bucket
	// bucketActions.CreateBucket(bucketName)
	bucketActions.ListBuckets()

	// Uncomment and use localstack to test with.
	// keyPrefix := []string{"deployhash112", "deploy234sfh", "deployTest321", "dep348dh", "d348hdfzui78"}
	// keySuffix := []string{"/index.html", "/css/font.css", "/image/hey.png"}
	// CreateTestObjects(bucketActions, keyPrefix, keySuffix)

	// Lists the objects from the bucket
	objects, err := bucketActions.ListObjects(bucketName)
	if err != nil {
		panic(err)
	}
	keyMap := CreateHashMap(objects)
	keyMapSorted := SortHashMap(keyMap)
	recentDeployments := FindMostRecentDeployments(keyMapSorted, *numbPtr)
	objectsToDelete := PopulateObjectsToDelete(objects, recentDeployments)

	if len(objectsToDelete) == 0 {
		log.Fatalf("No objects to delete detected.\nExiting...")
	}
	log.Printf("Found %v objects to delete.\n", len(objectsToDelete))
	for _, v := range objectsToDelete {
		log.Printf("Marked for deletion: %v\t\n", v)
	}

	if *toDelete {
		err = bucketActions.DeleteObjects(bucketName, objectsToDelete)
		if err != nil {
			panic(err)
		}
		// Show remaining objects
		log.Println("Remaining objects in bucket:")

		remainingObjects, err := bucketActions.ListObjects(bucketName)
		if err != nil {
			panic(err)
		}
		for _, object := range remainingObjects {
			log.Printf("\t%v\n", *object.Key)
		}
	}
}
