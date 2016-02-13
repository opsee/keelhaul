package service

type j map[string]interface{}

var swaggerMap = j{
	"basePath": "/",
	"swagger":  "2.0",
	"info": j{
		"title":       "Keelhaul API",
		"version":     "0.0.1",
		"description": "API for bastion management",
	},
	"tags": []j{
		j{
			"name":        "vpcs",
			"description": "VPC API",
		},
		j{
			"name":        "bastions",
			"description": "Bastion API",
		},
	},
	"paths": j{
		"/vpcs/scan": j{
			"post": j{
				"tags": []string{
					"vpcs",
				},
				"operationId": "scanVPC",
				"summary":     "Scan a VPC",
				"parameters":  []string{},
				"responses": j{
					"200": j{
						"description": "Description was not specified",
					},
					"401": j{
						"description": "Description was not specified",
					},
				},
			},
		},
		"/vpcs/launch": j{
			"post": j{
				"tags": []string{
					"vpcs",
				},
				"operationId": "launchVPC",
				"summary":     "Launch a VPC",
				"parameters":  []string{},
				"responses": j{
					"200": j{
						"description": "Description was not specified",
					},
					"401": j{
						"description": "Description was not specified",
					},
				},
			},
		},
		"/vpcs/cloudformation": j{
			"post": j{
				"tags": []string{
					"vpcs",
				},
				"operationId": "getCloudformation",
				"summary":     "Get Bastion Cloudformation Template",
				"parameters":  []string{},
				"responses": j{
					"200": j{
						"description": "Description was not specified",
					},
					"401": j{
						"description": "Description was not specified",
					},
				},
			},
		},

		// "/bastions": j{
		// 	"get": j{
		// 		"tags": []string{
		// 			"bastions",
		// 		},
		// 		"operationId": "listBastions",
		// 		"summary":     "Lists bastions.",
		// 		"parameters":  []string{},
		// 		"responses": j{
		// 			"200": j{
		// 				"description": "Description was not specified",
		// 			},
		// 			"401": j{
		// 				"description": "Description was not specified",
		// 			},
		// 		},
		// 	},
		// },
	},
	"definitions": j{},
	"consumes":    j{},
	"produces":    j{},
}
