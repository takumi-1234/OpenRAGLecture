// open-rag-lecture/internal/batch/processor/chunk_processor.go
package processor

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	// ★★★ UniPDFのコアパッケージ(model)と抽出パッケージ(extractor)をインポート ★★★
	"github.com/takumi-1234/OpenRAGLecture/internal/domain/model"
	"github.com/unidoc/unipdf/v3/extractor"
	unipdfmodel "github.com/unidoc/unipdf/v3/model"
)

const (
	defaultChunkSize   = 1000 // characters
	defaultOverlapSize = 200  // characters
)

// ChunkProcessor defines the interface for processing a document into chunks.
type ChunkProcessor interface {
	Process(ctx context.Context, doc *model.Document, fileContent []byte) ([]*model.Page, []*model.Chunk, error)
}

// pdfChunkProcessor is an implementation of ChunkProcessor for PDF files.
type pdfChunkProcessor struct {
	chunkSize             int
	overlapSize           int
	embeddingModelVersion string
}

// NewPDFChunkProcessor creates a new processor for PDF files.
func NewPDFChunkProcessor(embeddingModelVersion string) ChunkProcessor {
	return &pdfChunkProcessor{
		chunkSize:             defaultChunkSize,
		overlapSize:           defaultOverlapSize,
		embeddingModelVersion: embeddingModelVersion,
	}
}

// Process extracts text from a PDF, creates page and chunk models.
func (p *pdfChunkProcessor) Process(ctx context.Context, doc *model.Document, fileContent []byte) ([]*model.Page, []*model.Chunk, error) {
	pagesContent, err := p.extractTextFromPDF(fileContent)
	if err != nil {
		// PDFからのテキスト抽出に失敗した場合、ファイル全体をプレーンテキストとして扱うフォールバック処理
		log.Printf("Could not process file for doc %d as PDF with UniPDF, falling back to plain text: %v", doc.ID, err)
		pagesContent = map[int]string{1: string(fileContent)}
	}

	var pages []*model.Page
	var chunks []*model.Chunk
	var globalChunkIndex int

	for pageNum, pageText := range pagesContent {
		// 空のページはスキップ
		if strings.TrimSpace(pageText) == "" {
			continue
		}

		page := &model.Page{
			DocumentID: doc.ID,
			PageNumber: pageNum,
			Text:       pageText,
		}
		pages = append(pages, page)

		pageChunks := p.splitTextIntoChunks(pageText)

		for _, chunkText := range pageChunks {
			chunk := &model.Chunk{
				DocumentID:            doc.ID,
				CourseID:              doc.CourseID,
				SemesterID:            doc.SemesterID,
				ChunkIndex:            globalChunkIndex,
				Text:                  chunkText,
				EmbeddingID:           uuid.New().String(),
				EmbeddingModelVersion: p.embeddingModelVersion,
			}
			chunks = append(chunks, chunk)
			globalChunkIndex++
		}
	}

	return pages, chunks, nil
}

// ★★★ UniPDFのAPIを使用してテキスト抽出ロジックを実装 ★★★
// extractTextFromPDFは、UniPDFライブラリを使用してPDFのバイトスライスからページごとのテキストを抽出します。
func (p *pdfChunkProcessor) extractTextFromPDF(data []byte) (map[int]string, error) {
	// バイトスライスからPDFリーダーを作成します。
	pdfReader, err := unipdfmodel.NewPdfReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create new pdf reader with UniPDF: %w", err)
	}

	// PDFの総ページ数を取得します。
	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return nil, fmt.Errorf("failed to get number of pages with UniPDF: %w", err)
	}

	textByPage := make(map[int]string, numPages)
	for i := 1; i <= numPages; i++ {
		// ページ番号を指定してページオブジェクトを取得します。
		page, err := pdfReader.GetPage(i)
		if err != nil {
			return nil, fmt.Errorf("failed to get page %d with UniPDF: %w", i, err)
		}

		// ページからテキスト抽出用のエクストラクタを作成します。
		ex, err := extractor.New(page)
		if err != nil {
			return nil, fmt.Errorf("failed to create extractor for page %d with UniPDF: %w", i, err)
		}

		// ページ全体のテキストを抽出します。
		text, err := ex.ExtractText()
		if err != nil {
			return nil, fmt.Errorf("failed to extract text from page %d with UniPDF: %w", i, err)
		}
		textByPage[i] = text
	}

	return textByPage, nil
}

// splitTextIntoChunksは、指定されたテキストを固定サイズのチャンクに分割します。
// チャンク間にはオーバーラップを持たせることができます。
func (p *pdfChunkProcessor) splitTextIntoChunks(text string) []string {
	if len(text) <= p.chunkSize {
		return []string{text}
	}

	var chunks []string
	start := 0
	for start < len(text) {
		end := start + p.chunkSize
		if end > len(text) {
			end = len(text)
		}
		chunks = append(chunks, text[start:end])

		// 次のチャンクの開始位置を、オーバーラップを考慮して設定します。
		start += p.chunkSize - p.overlapSize
		if start >= len(text) {
			break
		}
	}
	return chunks
}
