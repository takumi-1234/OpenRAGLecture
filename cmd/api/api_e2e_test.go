// open-rag-lecture/cmd/api/api_e2e_test.go

package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
	"github.com/takumi-1234/OpenRAGLecture/internal/domain/model"
	"github.com/takumi-1234/OpenRAGLecture/internal/interface/repository/mysql"
	"github.com/takumi-1234/OpenRAGLecture/internal/usecase/output"
	"github.com/takumi-1234/OpenRAGLecture/pkg/config"
	"gorm.io/gorm"
)

// E2ETestSuite は、E2Eテスト全体のスイートを定義します。
type E2ETestSuite struct {
	suite.Suite
	db      *gorm.DB
	cfg     config.Config
	client  *http.Client
	baseURL string

	// テストスイート全体で共有する基本データ
	instructorEmail    string
	instructorPassword string
	courseID           uint64

	// シナリオテストで引き継ぐデータ
	studentEmail    string
	studentPassword string
	studentUserID   uint64
	studentToken    string
	instructorToken string
	uploadedDocID   uint64
}

// SetupSuite は、テストスイートの実行前に一度だけ実行されます。
func (s *E2ETestSuite) SetupSuite() {
	err := godotenv.Load("../../.env")
	s.Require().NoError(err, "Failed to load .env file")

	cfg, err := config.LoadConfig("../../configs")
	s.Require().NoError(err, "Failed to load config")
	cfg.Database.MySQL.Host = "localhost"
	hostPort := os.Getenv("MYSQL_HOST_PORT")
	if hostPort == "" {
		hostPort = "3406"
	}
	cfg.Database.MySQL.Port = hostPort
	s.cfg = cfg
	s.baseURL = fmt.Sprintf("http://localhost:%s", s.cfg.Server.Port)
	s.client = &http.Client{Timeout: 10 * time.Second}

	db, err := mysql.NewGORMClient(s.cfg.Database.MySQL)
	s.Require().NoError(err, "Failed to connect to DB")
	s.db = db

	// DBマイグレーションが完了したことをポーリングで確認します。
	// これにより、テストの安定性がさらに向上します。
	log.Println("Waiting for database migration to be reflected...")
	s.Require().Eventually(func() bool {
		// マイグレーションによって作成されるはずのテーブルのうち、代表的なもの(e.g., users)が存在するかを確認
		hasTable := s.db.Migrator().HasTable(&model.User{})
		return hasTable
	}, 10*time.Second, 500*time.Millisecond, "users table not found after migration")
	log.Println("Database is ready.")

	s.cleanupTestData()
	s.setupInitialData()
}

// TearDownSuite は、すべてのテストが完了した後に一度だけ実行されます。
func (s *E2ETestSuite) TearDownSuite() {
	log.Println("Tearing down suite, cleaning up all test data...")
	s.cleanupTestData()
	sqlDB, _ := s.db.DB()
	sqlDB.Close()
}

// cleanupTestData は、テーブルのデータを削除する安全な方法です。
func (s *E2ETestSuite) cleanupTestData() {
	log.Println("Cleaning up test data from tables...")
	s.db.Unscoped().Exec("SET FOREIGN_KEY_CHECKS = 0;")
	s.db.Unscoped().Exec("DELETE FROM feedbacks")
	s.db.Unscoped().Exec("DELETE FROM answer_sources")
	s.db.Unscoped().Exec("DELETE FROM answers")
	s.db.Unscoped().Exec("DELETE FROM questions")
	s.db.Unscoped().Exec("DELETE FROM chunks")
	s.db.Unscoped().Exec("DELETE FROM pages")
	s.db.Unscoped().Exec("DELETE FROM documents")
	s.db.Unscoped().Exec("DELETE FROM enrollments")
	s.db.Unscoped().Exec("DELETE FROM courses")
	s.db.Unscoped().Exec("DELETE FROM users")
	s.db.Unscoped().Exec("DELETE FROM semesters")
	s.db.Unscoped().Exec("SET FOREIGN_KEY_CHECKS = 1;")
}

// setupInitialData は、テストスイートで共通して使用する初期データを準備します。
func (s *E2ETestSuite) setupInitialData() {
	log.Println("Setting up initial test data (semester, instructor, course)...")
	s.instructorEmail = "instructor.e2e@example.com"
	s.instructorPassword = "instructorpass123"

	semester := model.Semester{
		Name:      "Test Semester for E2E",
		StartDate: time.Now(),
		EndDate:   time.Now().AddDate(0, 4, 0),
	}
	s.Require().NoError(s.db.Create(&semester).Error)

	instructor := model.User{
		Email:       s.instructorEmail,
		DisplayName: sql.NullString{String: "E2E Instructor", Valid: true},
		Role:        model.RoleInstructor,
	}
	err := instructor.SetPassword(s.instructorPassword)
	s.Require().NoError(err, "Failed to set password for instructor")
	s.Require().NoError(s.db.Create(&instructor).Error)
	log.Printf("Test instructor created with ID: %d", instructor.ID)

	course := model.Course{
		Code:         "E2E-TEST101",
		Title:        "Test Course for E2E",
		SemesterID:   semester.ID,
		InstructorID: instructor.ID,
	}
	s.Require().NoError(s.db.Create(&course).Error)
	s.courseID = course.ID
	log.Printf("Test course created with ID: %d", s.courseID)
}

// TestE2ETestSuite は、Testifyスイートを実行するためのエントリーポイントです。
func TestE2ETestSuite(t *testing.T) {
	suite.Run(t, new(E2ETestSuite))
}

// TestA_AuthFlow は認証関連のAPIフローをテストします。
func (s *E2ETestSuite) TestA_AuthFlow() {
	s.studentEmail = fmt.Sprintf("student_%d_e2e@example.com", time.Now().UnixNano())
	s.studentPassword = "studentpassword123"

	s.Run("Success_RegisterStudent", func() {
		body := map[string]string{"email": s.studentEmail, "password": s.studentPassword}
		jsonBody, _ := json.Marshal(body)
		resp, err := http.Post(s.baseURL+"/auth/register", "application/json", bytes.NewBuffer(jsonBody))
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Require().Equal(http.StatusCreated, resp.StatusCode)
	})

	s.Run("Failure_RegisterDuplicateStudent", func() {
		body := map[string]string{"email": s.studentEmail, "password": s.studentPassword}
		jsonBody, _ := json.Marshal(body)
		resp, err := http.Post(s.baseURL+"/auth/register", "application/json", bytes.NewBuffer(jsonBody))
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Require().Equal(http.StatusConflict, resp.StatusCode)
	})

	s.Run("Success_LoginStudent", func() {
		var loginResponse output.LoginOutput
		s.login(s.studentEmail, s.studentPassword, &loginResponse)
		s.NotEmpty(loginResponse.AccessToken, "Access token should not be empty for student")
		s.studentToken = loginResponse.AccessToken

		// ログインが成功したことをもってユーザーの存在を保証し、DBからIDを取得する
		var user model.User
		err := s.db.Where("email = ?", s.studentEmail).First(&user).Error
		s.Require().NoError(err, "Logged in user should exist in DB")
		s.studentUserID = user.ID
		s.Require().NotZero(s.studentUserID, "Student User ID should not be zero")
	})

	s.Run("Failure_LoginInvalidPassword", func() {
		body := map[string]string{"email": s.studentEmail, "password": "wrongpassword"}
		jsonBody, _ := json.Marshal(body)
		resp, err := http.Post(s.baseURL+"/auth/login", "application/json", bytes.NewBuffer(jsonBody))
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Require().Equal(http.StatusUnauthorized, resp.StatusCode, "Login with wrong password should fail")
	})

	s.Run("Success_LoginInstructor", func() {
		var loginResponse output.LoginOutput
		s.login(s.instructorEmail, s.instructorPassword, &loginResponse)
		s.NotEmpty(loginResponse.AccessToken, "Access token should not be empty for instructor")
		s.instructorToken = loginResponse.AccessToken
	})
}

// TestB_CourseEnrollment は講義への履修登録フローをテストします。
func (s *E2ETestSuite) TestB_CourseEnrollment() {
	s.Require().NotEmpty(s.studentToken, "Student must be logged in to enroll")

	s.Run("Success_EnrollStudent", func() {
		url := fmt.Sprintf("%s/api/courses/%d/enrollments", s.baseURL, s.courseID)
		req, _ := http.NewRequest("POST", url, nil)
		req.Header.Set("Authorization", "Bearer "+s.studentToken)
		resp, err := s.client.Do(req)
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Require().Equal(http.StatusCreated, resp.StatusCode)
	})

	s.Run("Failure_EnrollStudentTwice", func() {
		url := fmt.Sprintf("%s/api/courses/%d/enrollments", s.baseURL, s.courseID)
		req, _ := http.NewRequest("POST", url, nil)
		req.Header.Set("Authorization", "Bearer "+s.studentToken)
		resp, err := s.client.Do(req)
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Require().Equal(http.StatusConflict, resp.StatusCode)
	})
}

// TestC_FileUploadAndProcessing はファイルのアップロードとバッチ処理の完了をテストします。
func (s *E2ETestSuite) TestC_FileUploadAndProcessing() {
	s.Require().NotEmpty(s.instructorToken, "Instructor must be logged in to upload files")

	dummyPDFPath, cleanup := s.createDummyPDF()
	defer cleanup()

	s.Run("Success_InstructorUploadsFile", func() {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("course_id", fmt.Sprintf("%d", s.courseID))
		part, _ := writer.CreateFormFile("file", filepath.Base(dummyPDFPath))
		file, _ := os.Open(dummyPDFPath)
		_, _ = io.Copy(part, file)
		file.Close()
		writer.Close()

		req, _ := http.NewRequest("POST", s.baseURL+"/api/files/upload", body)
		req.Header.Set("Authorization", "Bearer "+s.instructorToken)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := s.client.Do(req)
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Require().Equal(http.StatusCreated, resp.StatusCode)

		var uploadResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&uploadResponse)
		s.Require().NoError(err)
		s.Contains(uploadResponse, "document_id")
		s.uploadedDocID = uint64(uploadResponse["document_id"].(float64))
	})

	s.Run("Success_BatchProcessingCompletes", func() {
		s.Require().NotZero(s.uploadedDocID, "Document ID must be set from upload step")

		log.Println("Executing 'sync-documents' batch job...")
		cmd := exec.Command("docker", "compose", "run", "--rm", "batch", "sync-documents")
		output, err := cmd.CombinedOutput()
		s.Require().NoError(err, "Batch job failed. Output:\n%s", string(output))
		log.Printf("Batch job finished. Output:\n%s", string(output))

		log.Println("Waiting for chunks to be created in DB...")
		s.Require().Eventually(func() bool {
			var count int64
			err := s.db.Model(&model.Chunk{}).Where("document_id = ?", s.uploadedDocID).Count(&count).Error
			if err != nil {
				return false
			}
			return count > 0
		}, 30*time.Second, 1*time.Second, "Chunks were not created for the document within the time limit")
		log.Println("Chunks created successfully!")
	})
}

// TestD_QAFlow はQA機能のAPIフローをテストします。
func (s *E2ETestSuite) TestD_QAFlow() {
	s.Require().NotEmpty(s.studentToken, "Student must be logged in for QA")
	s.Require().NotZero(s.uploadedDocID, "A document must have been processed for QA")

	s.Run("Success_EnrolledStudentAsksQuestion", func() {
		qaBody := map[string]interface{}{"course_id": s.courseID, "query": "What is the test content?"}
		jsonBody, _ := json.Marshal(qaBody)

		req, _ := http.NewRequest("POST", s.baseURL+"/api/qa/ask", bytes.NewBuffer(jsonBody))
		req.Header.Set("Authorization", "Bearer "+s.studentToken)
		req.Header.Set("Content-Type", "application/json")
		resp, err := s.client.Do(req)
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Require().Equal(http.StatusOK, resp.StatusCode, "QA request should return 200 OK")
		var qaResponse map[string]string
		err = json.NewDecoder(resp.Body).Decode(&qaResponse)
		s.Require().NoError(err)
		s.Contains(qaResponse, "answer")
		s.Contains(strings.ToLower(qaResponse["answer"]), "test content", "Answer should contain the text from the PDF")
		log.Printf("QA success. Answer: %s", qaResponse["answer"])
	})

	s.Run("Failure_AskAboutUnenrolledCourse", func() {
		invalidCourseID := uint64(99999)
		qaBody := map[string]interface{}{"course_id": invalidCourseID, "query": "Does not matter"}
		jsonBody, _ := json.Marshal(qaBody)

		req, _ := http.NewRequest("POST", s.baseURL+"/api/qa/ask", bytes.NewBuffer(jsonBody))
		req.Header.Set("Authorization", "Bearer "+s.studentToken)
		req.Header.Set("Content-Type", "application/json")
		resp, err := s.client.Do(req)
		s.Require().NoError(err)
		defer resp.Body.Close()
		s.Require().Equal(http.StatusForbidden, resp.StatusCode, "Should not be able to ask about an unenrolled course")
	})
}

// --- Helper Methods ---

func (s *E2ETestSuite) login(email, password string, target interface{}) {
	body := map[string]string{"email": email, "password": password}
	jsonBody, _ := json.Marshal(body)
	resp, err := http.Post(s.baseURL+"/auth/login", "application/json", bytes.NewBuffer(jsonBody))
	s.Require().NoError(err, "Login request failed")
	defer resp.Body.Close()
	s.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Login should be successful for %s", email))

	err = json.NewDecoder(resp.Body).Decode(target)
	s.Require().NoError(err, "Failed to decode login response")
}

func (s *E2ETestSuite) createDummyPDF() (path string, cleanup func()) {
	path = "dummy_e2e_test_file.pdf"
	content := []byte(
		`%PDF-1.7
1 0 obj << /Type /Catalog /Pages 2 0 R >> endobj
2 0 obj << /Type /Pages /Kids [3 0 R] /Count 1 >> endobj
3 0 obj << /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >> endobj
4 0 obj << /Length 44 >> stream
BT /F1 24 Tf 100 700 Td (This is the test content.) Tj ET
endstream endobj
5 0 obj << /Type /Font /Subtype /Type1 /BaseFont /Helvetica >> endobj
xref
0 6
0000000000 65535 f 
0000000010 00000 n 
0000000059 00000 n 
0000000118 00000 n 
0000000213 00000 n 
0000000282 00000 n 
trailer
<< /Size 6 /Root 1 0 R >>
startxref
358
%%EOF`)
	err := os.WriteFile(path, content, 0644)
	s.Require().NoError(err)
	cleanup = func() { os.Remove(path) }
	return path, cleanup
}
