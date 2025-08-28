// open-rag-lecture/internal/interface/repository/qdrant/vector_repository.go

package qdrant

import (
	"context"
	"fmt"
	"log"

	pb "github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/takumi-1234/OpenRAGLecture/internal/domain/model"
	"github.com/takumi-1234/OpenRAGLecture/internal/domain/repository"
	"github.com/takumi-1234/OpenRAGLecture/pkg/config"
)

type qdrantRepository struct {
	pointsClient      pb.PointsClient
	collectionsClient pb.CollectionsClient
	collectionName    string
	vectorSize        uint64
}

// NewQdrantRepository creates a new VectorRepository implementation for Qdrant.
func NewQdrantRepository(cfg config.QdrantConfig) (repository.VectorRepository, error) {
	var opts []grpc.DialOption
	if cfg.UseTLS {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(nil)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", cfg.Host, cfg.GrpcPort), opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to qdrant: %w", err)
	}

	return &qdrantRepository{
		pointsClient:      pb.NewPointsClient(conn),
		collectionsClient: pb.NewCollectionsClient(conn),
		collectionName:    cfg.CollectionName,
		vectorSize:        cfg.VectorSize,
	}, nil
}

func (r *qdrantRepository) Upsert(ctx context.Context, chunks []*model.Chunk, vectors [][]float32) error {
	points := make([]*pb.PointStruct, len(chunks))
	for i, chunk := range chunks {
		points[i] = &pb.PointStruct{
			Id:      &pb.PointId{PointIdOptions: &pb.PointId_Uuid{Uuid: chunk.EmbeddingID}},
			Vectors: &pb.Vectors{VectorsOptions: &pb.Vectors_Vector{Vector: &pb.Vector{Data: vectors[i]}}},
			Payload: map[string]*pb.Value{
				"text":      {Kind: &pb.Value_StringValue{StringValue: chunk.Text}},
				"chunk_id":  {Kind: &pb.Value_IntegerValue{IntegerValue: int64(chunk.ID)}},
				"doc_id":    {Kind: &pb.Value_IntegerValue{IntegerValue: int64(chunk.DocumentID)}},
				"course_id": {Kind: &pb.Value_IntegerValue{IntegerValue: int64(chunk.CourseID)}},
			},
		}
	}

	wait := true
	_, err := r.pointsClient.Upsert(ctx, &pb.UpsertPoints{
		CollectionName: r.collectionName,
		Wait:           &wait,
		Points:         points,
	})
	return err
}

func (r *qdrantRepository) Search(ctx context.Context, queryVector []float32, courseID uint64, limit int) ([]model.RetrievedChunk, error) {
	searchRequest := &pb.SearchPoints{
		CollectionName: r.collectionName,
		Vector:         queryVector,
		Limit:          uint64(limit),
		WithPayload:    &pb.WithPayloadSelector{SelectorOptions: &pb.WithPayloadSelector_Enable{Enable: true}},
		Filter: &pb.Filter{
			Must: []*pb.Condition{
				{
					ConditionOneOf: &pb.Condition_Field{
						Field: &pb.FieldCondition{
							Key: "course_id",
							Match: &pb.Match{
								MatchValue: &pb.Match_Integer{Integer: int64(courseID)},
							},
						},
					},
				},
			},
		},
	}

	res, err := r.pointsClient.Search(ctx, searchRequest)
	if err != nil {
		return nil, err
	}

	retrievedChunks := make([]model.RetrievedChunk, len(res.GetResult()))
	for i, point := range res.GetResult() {
		payload := point.GetPayload()
		retrievedChunks[i] = model.RetrievedChunk{
			Chunk: model.Chunk{
				Base:       model.Base{ID: uint64(payload["chunk_id"].GetIntegerValue())},
				DocumentID: uint64(payload["doc_id"].GetIntegerValue()),
				CourseID:   uint64(payload["course_id"].GetIntegerValue()),
				Text:       payload["text"].GetStringValue(),
			},
			Score: point.GetScore(),
		}
	}
	return retrievedChunks, nil
}

// RecreateCollection deletes and then creates the collection.
func (r *qdrantRepository) RecreateCollection(ctx context.Context) error {
	// 1. Delete the collection if it exists.
	_, err := r.collectionsClient.Delete(ctx, &pb.DeleteCollection{
		CollectionName: r.collectionName,
	})
	if err != nil && status.Code(err) != codes.NotFound {
		return fmt.Errorf("failed to delete qdrant collection: %w", err)
	}
	log.Printf("Qdrant collection '%s' deleted (if it existed).", r.collectionName)

	// 2. Create a new collection.
	_, err = r.collectionsClient.Create(ctx, &pb.CreateCollection{
		CollectionName: r.collectionName,
		VectorsConfig: &pb.VectorsConfig{
			Config: &pb.VectorsConfig_Params{
				Params: &pb.VectorParams{
					Size:     r.vectorSize,
					Distance: pb.Distance_Cosine,
				},
			},
		},
	})
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			log.Printf("Qdrant collection '%s' already exists (race condition handled).", r.collectionName)
			return nil
		}
		return fmt.Errorf("failed to create qdrant collection: %w", err)
	}
	log.Printf("Qdrant collection '%s' created successfully.", r.collectionName)
	return nil
}

// EnsureCollectionExists creates the collection only if it does not exist.
func (r *qdrantRepository) EnsureCollectionExists(ctx context.Context) error {
	// 1. Check if the collection exists.
	res, err := r.collectionsClient.Get(ctx, &pb.GetCollectionInfoRequest{
		CollectionName: r.collectionName,
	})
	if err == nil && res != nil {
		log.Printf("Qdrant collection '%s' already exists. Skipping creation.", r.collectionName)
		return nil // Collection already exists
	}

	if status.Code(err) != codes.NotFound {
		return fmt.Errorf("failed to check qdrant collection existence: %w", err)
	}

	// 2. If not found, create it.
	log.Printf("Qdrant collection '%s' not found. Creating...", r.collectionName)
	_, err = r.collectionsClient.Create(ctx, &pb.CreateCollection{
		CollectionName: r.collectionName,
		VectorsConfig: &pb.VectorsConfig{
			Config: &pb.VectorsConfig_Params{
				Params: &pb.VectorParams{
					Size:     r.vectorSize,
					Distance: pb.Distance_Cosine,
				},
			},
		},
	})
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			return nil
		}
		return fmt.Errorf("failed to create qdrant collection: %w", err)
	}

	log.Printf("Qdrant collection '%s' created successfully.", r.collectionName)
	return nil
}
