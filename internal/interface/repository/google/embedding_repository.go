// OpenRAGLecture/internal/interface/repository/google/embedding_repository.go
package google

import (
	"context"
	"fmt"

	aiplatform "cloud.google.com/go/aiplatform/apiv1"
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"github.com/takumi-1234/OpenRAGLecture/internal/domain/repository"
	"github.com/takumi-1234/OpenRAGLecture/pkg/config"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/structpb"
)

type googleEmbeddingRepository struct {
	client   *aiplatform.PredictionClient
	endpoint string
}

// NewGoogleEmbeddingRepository creates a new EmbeddingRepository using the Vertex AI SDK.
func NewGoogleEmbeddingRepository(cfg config.GoogleConfig) (repository.EmbeddingRepository, error) {
	if cfg.ProjectID == "" || cfg.Location == "" {
		return nil, fmt.Errorf("Google ProjectID and Location must be configured for Vertex AI")
	}

	ctx := context.Background()
	apiEndpoint := fmt.Sprintf("%s-aiplatform.googleapis.com:443", cfg.Location)

	// ★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★
	// 修正点: APIキー認証を削除し、サービスアカウント認証 (ADC) に統一する
	// option.WithAPIKey() を削除することで、SDKは
	// GOOGLE_APPLICATION_CREDENTIALS 環境変数から認証情報を自動で読み込みます。
	// ★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★
	opts := []option.ClientOption{
		option.WithEndpoint(apiEndpoint),
	}

	client, err := aiplatform.NewPredictionClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Vertex AI prediction client: %w", err)
	}

	endpoint := fmt.Sprintf("projects/%s/locations/%s/publishers/google/models/%s",
		cfg.ProjectID, cfg.Location, cfg.EmbeddingModel)

	return &googleEmbeddingRepository{
		client:   client,
		endpoint: endpoint,
	}, nil
}

// CreateEmbeddings generates vector embeddings for a batch of texts.
func (r *googleEmbeddingRepository) CreateEmbeddings(ctx context.Context, texts []string, taskType string) ([][]float32, error) {
	instances := make([]*structpb.Value, len(texts))
	for i, text := range texts {
		instance, err := structpb.NewStruct(map[string]interface{}{
			"content":   text,
			"task_type": taskType,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create struct for instance %d: %w", i, err)
		}
		instances[i] = structpb.NewStructValue(instance)
	}

	// Parameters can be added here if needed, e.g., outputDimensionality
	// params, _ := structpb.NewStruct(map[string]interface{}{})

	req := &aiplatformpb.PredictRequest{
		Endpoint:  r.endpoint,
		Instances: instances,
		// Parameters: structpb.NewStructValue(params),
	}

	resp, err := r.client.Predict(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("prediction request failed: %w", err)
	}

	embeddings := make([][]float32, len(resp.Predictions))
	for i, prediction := range resp.Predictions {
		structVal := prediction.GetStructValue()
		if structVal == nil {
			return nil, fmt.Errorf("prediction %d is not a struct", i)
		}

		embeddingVal, ok := structVal.Fields["embeddings"]
		if !ok {
			return nil, fmt.Errorf("no 'embeddings' field in prediction %d", i)
		}

		embeddingStruct := embeddingVal.GetStructValue()
		if embeddingStruct == nil {
			return nil, fmt.Errorf("'embeddings' field in prediction %d is not a struct", i)
		}

		valuesVal, ok := embeddingStruct.Fields["values"]
		if !ok {
			return nil, fmt.Errorf("no 'values' field in embedding %d", i)
		}

		values := valuesVal.GetListValue().GetValues()
		embeddings[i] = make([]float32, len(values))
		for j, value := range values {
			embeddings[i][j] = float32(value.GetNumberValue())
		}
	}

	return embeddings, nil
}
