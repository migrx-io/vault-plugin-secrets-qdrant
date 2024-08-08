package qdrant

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"time"

	"encoding/base64"
	"github.com/google/uuid"

	"github.com/hashicorp/vault/sdk/logical"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pb "github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
)

const (
	SYS_ROLE_TABLE = "sys_roles"
)

type QdrantClient struct {
	//client *grpc.ClientConn
}

func (c *QdrantClient) createRole(ctx context.Context, s logical.Storage, role *RoleParameters) error {

	conn, err := getClientQdrant(ctx, s, role.DBId)

	if err != nil {
		return err
	}

	defer conn.Close()

	client := pb.NewCollectionsClient(conn)
	client_p := pb.NewPointsClient(conn) //PointsClient

	// Contact the server
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	isExists, err := checkExistCollection(ctx, client)

	if err != nil {
		return err
	}

	if !isExists {
		//create colection
		err = createNewCollection(ctx, client)

		if err != nil {
			return err
		}

	}

	// delete same key if exists
	err = deleteRolePoint(ctx, client_p, role.RoleId)
	if err != nil {
		return err
	}

	//add new role name
	err = createRolePoint(ctx, client_p, role.RoleId)
	if err != nil {
		return err
	}

	return nil

}

func (c *QdrantClient) deleteRole(ctx context.Context, s logical.Storage, role *RoleParameters) error {

	conn, err := getClientQdrant(ctx, s, role.DBId)

	if err != nil {
		return err
	}

	defer conn.Close()

	client := pb.NewCollectionsClient(conn)
	client_p := pb.NewPointsClient(conn) //PointsClient

	// Contact the server
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	isExists, err := checkExistCollection(ctx, client)

	if err != nil {
		return err
	}

	if isExists {
		//delete point
		err = deleteRolePoint(ctx, client_p, role.RoleId)

		if err != nil {
			return err
		}

	}

	return nil

}

func loadTLSCredentials(isSecure bool, CA string) (credentials.TransportCredentials, error) {

	if !isSecure {
		return insecure.NewCredentials(), nil
	}

	if CA == "" {
		return credentials.NewTLS(&tls.Config{}), nil
	}

	// Load certificate of the CA who signed server's certificate
	pemServerCA, err := base64.StdEncoding.DecodeString(CA)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	// Create the credentials and return it
	config := &tls.Config{
		RootCAs: certPool,
	}

	return credentials.NewTLS(config), nil
}

func deleteRolePoint(ctx context.Context, client pb.PointsClient, name string) error {

	// delete role index for sys_roles

	// filtered search
	_, err := client.Delete(ctx, &pb.DeletePoints{
		CollectionName: "sys_roles",
		Points: &pb.PointsSelector{
			PointsSelectorOneOf: &pb.PointsSelector_Filter{

				&pb.Filter{
					Should: []*pb.Condition{
						{
							ConditionOneOf: &pb.Condition_Field{
								Field: &pb.FieldCondition{
									Key: "role",
									Match: &pb.Match{
										MatchValue: &pb.Match_Keyword{
											Keyword: name,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})

	if err != nil {
		return nil
	}

	return nil

}

func createRolePoint(ctx context.Context, client pb.PointsClient, name string) error {

	// create role index for sys_roles
	// Create keyword field index
	fieldIndex1Type := pb.FieldType_FieldTypeKeyword
	fieldIndex1Name := "role"
	_, err := client.CreateFieldIndex(ctx, &pb.CreateFieldIndexCollection{
		CollectionName: SYS_ROLE_TABLE,
		FieldName:      fieldIndex1Name,
		FieldType:      &fieldIndex1Type,
	})

	if err != nil {
		return err
	}

	// create points and insert
	// Upsert points
	waitUpsert := true
	upsertPoints := []*pb.PointStruct{
		{
			// Point Id is number or UUID
			Id: &pb.PointId{
				PointIdOptions: &pb.PointId_Uuid{Uuid: uuid.New().String()},
			},
			Vectors: &pb.Vectors{VectorsOptions: &pb.Vectors_Vector{Vector: &pb.Vector{Data: []float32{0.1}}}},
			Payload: map[string]*pb.Value{
				"role": {
					Kind: &pb.Value_StringValue{StringValue: name},
				},
			},
		},
	}

	_, err = client.Upsert(ctx, &pb.UpsertPoints{
		CollectionName: SYS_ROLE_TABLE,
		Wait:           &waitUpsert,
		Points:         upsertPoints,
	})
	if err != nil {
		return err
	}

	return nil

}

func createNewCollection(ctx context.Context, client pb.CollectionsClient) error {

	// Create new collection
	//var defaultSegmentNumber uint64 = 2
	var onDisk = true
	_, err := client.Create(ctx, &pb.CreateCollection{
		CollectionName: SYS_ROLE_TABLE,
		VectorsConfig: &pb.VectorsConfig{Config: &pb.VectorsConfig_Params{
			Params: &pb.VectorParams{
				Size:     1,
				Distance: pb.Distance_Dot,
				OnDisk:   &onDisk,
			},
		}},
		//OptimizersConfig: &pb.OptimizersConfigDiff{
		//    DefaultSegmentNumber: &defaultSegmentNumber,
		//},
	})

	if err != nil {
		return err
	}

	return nil

}

func checkExistCollection(ctx context.Context, client pb.CollectionsClient) (bool, error) {

	resp, err := client.CollectionExists(ctx, &pb.CollectionExistsRequest{
		CollectionName: SYS_ROLE_TABLE,
	})

	if err != nil {
		return false, fmt.Errorf("Could not get collection: %v", err)

	}

	return resp.Result.Exists, nil

}

func getClientQdrant(ctx context.Context, s logical.Storage, dbId string) (*grpc.ClientConn, error) {

	// get stored signing keys
	config, err := readConfig(ctx, s, dbId)
	if err != nil {
		return nil, err
	}

	isTLS := config.TLS

	tlsCredentials, err := loadTLSCredentials(isTLS, "")
	if err != nil {
		return nil, err
	}

	interceptor := interceptorBuilder(config.SignKey)

	conn, err := grpc.NewClient(config.URL,
		grpc.WithTransportCredentials(tlsCredentials),
		grpc.WithUnaryInterceptor(interceptor))

	if err != nil {
		return nil, err
	}
	//defer conn.Close()

	return conn, nil
}

func interceptorBuilder(apiKey string) func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

	f := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		newCtx := metadata.AppendToOutgoingContext(ctx, "api-key", apiKey)
		return invoker(newCtx, method, req, reply, cc, opts...)
	}

	return f

}
