package handler

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/go-chi/chi/v5"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// 1. 创建一个数据库的模拟对象
 type MockDB struct {
	mock.Mock
}
// 2. 为模拟对象实现我们定义的接口
func (m *MockDB) CreateTaskForAgent(ctx context.Context, agentID int, taskType string) error {
	args := m.Called(ctx, agentID, taskType)
	return args.Error(0)
}

// 3. 编写测试函数
func TestTaskHandler_Create_Success(t *testing.T) {
	mockDB := new(MockDB)
	mockDB.On("CreateTaskForAgent", mock.Anything, 1, "DUMP_WECHAT_DATA").Return(nil)
	handler := TaskHandler{DB: mockDB}
	req := httptest.NewRequest("POST", "/v1/agents/1/tasks", nil)
	rr := httptest.NewRecorder()
	router := chi.NewRouter()
	router.Post("/v1/agents/{agentID}/tasks", handler.Create)
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)
	mockDB.AssertExpectations(t)
}

func TestTaskHandler_Create_DBError(t *testing.T) {
	mockDB := new(MockDB)
	mockDB.On("CreateTaskForAgent", mock.Anything, 2, "DUMP_WECHAT_DATA").Return(assert.AnError)
	handler := TaskHandler{DB: mockDB}
	req := httptest.NewRequest("POST", "/v1/agents/2/tasks", nil)
	rr := httptest.NewRecorder()
	router := chi.NewRouter()
	router.Post("/v1/agents/{agentID}/tasks", handler.Create)
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockDB.AssertExpectations(t)
}
