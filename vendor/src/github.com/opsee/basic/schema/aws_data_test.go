package schema

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/graphql-go/graphql"
	// autoscaling "github.com/opsee/basic/schema/aws/autoscaling"
	ec2 "github.com/opsee/basic/schema/aws/ec2"
	// elb "github.com/opsee/basic/schema/aws/elb"
	// rds "github.com/opsee/basic/schema/aws/rds"
	// googleproto "github.com/opsee/protobuf/proto/google/protobuf"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSchema(t *testing.T) {
	instance := &Instance{
		Id:         "i-666",
		CustomerId: "666-beef",
		Type:       "ec2",
		Resource: &Instance_Instance{
			Instance: &ec2.Instance{
				InstanceId:       aws.String("i-666"),
				InstanceType:     aws.String("t2.micro"),
				PrivateIpAddress: aws.String("666.666.666.666"),
				State: &ec2.InstanceState{
					Name: aws.String("running"),
				},
				Tags: []*ec2.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String("beast server"),
					},
				},
			},
		},
	}

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"instance": &graphql.Field{
					Type: GraphQLInstanceType,
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						return instance, nil
					},
				},
			},
		}),
	})

	if err != nil {
		t.Fatal(err)
	}

	queryResponse := graphql.Do(graphql.Params{Schema: schema, RequestString: `query instanceQuery {
		instance {
			id
			customer_id
			type
			resource {
				... on ec2Instance {
					InstanceId
					InstanceType
					PrivateIpAddress
					State {
						Name
					}
					Tags {
						Key
						Value
					}
				}
			}
		}
	}`})

	if queryResponse.HasErrors() {
		t.Fatalf("graphql query errors: %#v\n", queryResponse.Errors)
	}

	assert.EqualValues(t, instance.Id, getProp(queryResponse.Data, "instance", "id"))
}

func getProp(i interface{}, path ...interface{}) interface{} {
	cur := i

	for _, s := range path {
		switch cur.(type) {
		case map[string]interface{}:
			cur = cur.(map[string]interface{})[s.(string)]
			continue
		case []interface{}:
			cur = cur.([]interface{})[s.(int)]
			continue
		default:
			return cur
		}
	}

	return cur
}
