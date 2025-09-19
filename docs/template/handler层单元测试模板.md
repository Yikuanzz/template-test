# Handler层单元测试模板

## 概述

这是一个基于 [uber-go/mock](https://github.com/uber-go/mock) 和 Gin 框架的 Go 语言 handler 层单元测试模板，用于测试 HTTP 处理器层的业务逻辑。

## 核心特性

- ✅ 使用 gomock 生成接口 mock
- ✅ 基于 Gin 框架的 HTTP 测试
- ✅ 完整的请求/响应测试覆盖
- ✅ 参数验证和错误处理测试
- ✅ 业务逻辑隔离测试
- ✅ 支持中文和国际化测试

## 依赖安装

### 1. 安装 gomock 工具

```bash
go install go.uber.org/mock/mockgen@latest
```

### 2. 安装测试依赖

```bash
go get github.com/gin-gonic/gin
go get github.com/stretchr/testify/assert
go get go.uber.org/mock/mockgen
```

## 完整代码模板

### 1. 基础结构和导入

```go
package fast_answer

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"ticketing-system/db/model"
	"ticketing-system/utils/common"

	zhlog "codeup.aliyun.com/chevalierteam/zhanhai-kit/plugins/logger/zap"
)
```

### 2. 生成 Mock 接口

首先，为你的仓储接口生成 mock：

```bash
# 在项目根目录执行
mockgen -source=app/ticketing/client/internal/handler/fast_answer/fast_answer.go -destination=app/ticketing/client/internal/handler/fast_answer/mock_fast_answer_repo.go -package=fast_answer
```

### 3. 测试辅助函数

```go
// setupTestRouter 设置测试路由
func setupTestRouter(handler *FastAnswerHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// 注册路由
	router.GET("/api/fast-answers", handler.GetFastAnswers)
	router.GET("/api/fast-answers/:fast_answer_uid", handler.GetFastAnswerByUID)
	router.GET("/api/fast-answers/types", handler.GetFastAnswerTypes)
	
	return router
}

// createTestHandler 创建测试用的处理器
func createTestHandler(ctrl *gomock.Controller) (*FastAnswerHandler, *MockFastAnswerRepo) {
	mockRepo := NewMockFastAnswerRepo(ctrl)
	logger := zhlog.NewHelper(zhlog.NewLogger())
	handler := NewFastAnswerHandler(mockRepo, logger)
	return handler, mockRepo
}

// createTestFastAnswer 创建测试数据
func createTestFastAnswer(uid, content, answerType string) *model.FastAnswer {
	return &model.FastAnswer{
		FastAnswerUID: uid,
		Content:       content,
		Type:          answerType,
		FileURL:       "",
		ExtraData:     []byte(`{"test": true}`),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// assertJSONResponse 断言JSON响应
func assertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedMessage string) {
	t.Helper()
	
	assert.Equal(t, expectedStatus, w.Code)
	
	var response common.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedMessage, response.Message)
}
```

### 4. 完整测试示例

```go
func TestFastAnswerHandler_GetFastAnswers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler, mockRepo := createTestHandler(ctrl)
	router := setupTestRouter(handler)

	t.Run("获取所有快捷回复成功", func(t *testing.T) {
		// 准备测试数据
		testData := []*model.FastAnswer{
			createTestFastAnswer("test-1", "测试内容1", "text"),
			createTestFastAnswer("test-2", "测试内容2", "image"),
		}

		// 设置 mock 期望
		mockRepo.EXPECT().
			GetAll(gomock.Any()).
			Return(testData, nil).
			Times(1)

		// 执行请求
		req, _ := http.NewRequest("GET", "/api/fast-answers", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 验证响应
		assertJSONResponse(t, w, http.StatusOK, "获取成功")

		var response common.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// 验证响应数据结构
		respData, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.Contains(t, respData, "fast_answers")
		assert.Contains(t, respData, "total_count")
		assert.Equal(t, float64(2), respData["total_count"])
	})

	t.Run("按类型获取快捷回复成功", func(t *testing.T) {
		// 准备测试数据
		testData := []*model.FastAnswer{
			createTestFastAnswer("test-1", "文本内容", "text"),
		}

		// 设置 mock 期望
		mockRepo.EXPECT().
			GetByType(gomock.Any(), "text").
			Return(testData, nil).
			Times(1)

		// 执行请求
		req, _ := http.NewRequest("GET", "/api/fast-answers?type=text", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 验证响应
		assertJSONResponse(t, w, http.StatusOK, "获取成功")
	})

	t.Run("获取快捷回复失败", func(t *testing.T) {
		// 设置 mock 期望 - 返回错误
		mockRepo.EXPECT().
			GetAll(gomock.Any()).
			Return(nil, assert.AnError).
			Times(1)

		// 执行请求
		req, _ := http.NewRequest("GET", "/api/fast-answers", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 验证响应
		assertJSONResponse(t, w, http.StatusInternalServerError, "获取失败")
	})

	t.Run("空数据时返回空列表", func(t *testing.T) {
		// 设置 mock 期望 - 返回空数据
		mockRepo.EXPECT().
			GetAll(gomock.Any()).
			Return([]*model.FastAnswer{}, nil).
			Times(1)

		// 执行请求
		req, _ := http.NewRequest("GET", "/api/fast-answers", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 验证响应
		assertJSONResponse(t, w, http.StatusOK, "获取成功")

		var response common.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		respData, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, float64(0), respData["total_count"])
	})
}

func TestFastAnswerHandler_GetFastAnswerByUID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler, mockRepo := createTestHandler(ctrl)
	router := setupTestRouter(handler)

	t.Run("根据UID获取快捷回复成功", func(t *testing.T) {
		// 准备测试数据
		testData := createTestFastAnswer("test-uid", "测试内容", "text")

		// 设置 mock 期望
		mockRepo.EXPECT().
			GetByUID(gomock.Any(), "test-uid").
			Return(testData, nil).
			Times(1)

		// 执行请求
		req, _ := http.NewRequest("GET", "/api/fast-answers/test-uid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 验证响应
		assertJSONResponse(t, w, http.StatusOK, "获取成功")

		var response common.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// 验证响应数据
		respData, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "test-uid", respData["fast_answer_uid"])
		assert.Equal(t, "测试内容", respData["content"])
		assert.Equal(t, "text", respData["type"])
	})

	t.Run("快捷回复不存在", func(t *testing.T) {
		// 设置 mock 期望 - 返回错误
		mockRepo.EXPECT().
			GetByUID(gomock.Any(), "nonexistent-uid").
			Return(nil, assert.AnError).
			Times(1)

		// 执行请求
		req, _ := http.NewRequest("GET", "/api/fast-answers/nonexistent-uid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 验证响应
		assertJSONResponse(t, w, http.StatusNotFound, "快捷回复不存在")
	})

	t.Run("参数错误", func(t *testing.T) {
		// 执行请求 - 缺少必需参数
		req, _ := http.NewRequest("GET", "/api/fast-answers/", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 验证响应
		assert.Equal(t, http.StatusNotFound, w.Code) // Gin 路由不匹配
	})
}

func TestFastAnswerHandler_GetFastAnswerTypes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler, _ := createTestHandler(ctrl)
	router := setupTestRouter(handler)

	t.Run("获取快捷回复类型成功", func(t *testing.T) {
		// 执行请求
		req, _ := http.NewRequest("GET", "/api/fast-answers/types", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 验证响应
		assertJSONResponse(t, w, http.StatusOK, "获取成功")

		var response common.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// 验证响应数据结构
		respData, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.Contains(t, respData, "types")
		assert.Contains(t, respData, "count")
		assert.Equal(t, float64(3), respData["count"])
	})
}
```

### 5. 参数验证测试

```go
func TestFastAnswerHandler_ParameterValidation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler, _ := createTestHandler(ctrl)
	router := setupTestRouter(handler)

	t.Run("无效的查询参数", func(t *testing.T) {
		// 测试无效的查询参数格式
		req, _ := http.NewRequest("GET", "/api/fast-answers?type=invalid%20type%20with%20spaces", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 根据实际业务逻辑验证响应
		// 这里假设无效类型会被接受，实际测试中需要根据业务规则调整
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("空查询参数", func(t *testing.T) {
		// 测试空查询参数
		req, _ := http.NewRequest("GET", "/api/fast-answers?type=", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 应该正常处理空参数
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
```

### 6. 并发测试

```go
func TestFastAnswerHandler_Concurrency(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler, mockRepo := createTestHandler(ctrl)
	router := setupTestRouter(handler)

	t.Run("并发请求测试", func(t *testing.T) {
		// 准备测试数据
		testData := []*model.FastAnswer{
			createTestFastAnswer("test-1", "测试内容1", "text"),
		}

		// 设置 mock 期望 - 允许多次调用
		mockRepo.EXPECT().
			GetAll(gomock.Any()).
			Return(testData, nil).
			AnyTimes()

		// 并发执行多个请求
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				defer func() { done <- true }()
				
				req, _ := http.NewRequest("GET", "/api/fast-answers", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				
				assert.Equal(t, http.StatusOK, w.Code)
			}()
		}

		// 等待所有请求完成
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}
```

### 7. 性能测试

```go
func BenchmarkFastAnswerHandler_GetFastAnswers(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	handler, mockRepo := createTestHandler(ctrl)
	router := setupTestRouter(handler)

	// 准备测试数据
	testData := []*model.FastAnswer{
		createTestFastAnswer("test-1", "测试内容1", "text"),
		createTestFastAnswer("test-2", "测试内容2", "image"),
	}

	// 设置 mock 期望
	mockRepo.EXPECT().
		GetAll(gomock.Any()).
		Return(testData, nil).
		AnyTimes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/fast-answers", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
```

## 关键配置说明

### 1. Mock 生成
- 使用 `mockgen` 工具生成接口 mock
- 支持类型安全的 mock 方法
- 自动生成期望设置和验证

### 2. Gin 测试模式
- 设置 `gin.TestMode` 减少日志输出
- 使用 `httptest.NewRecorder()` 捕获响应
- 支持完整的 HTTP 请求/响应测试

### 3. 测试数据管理
- 使用工厂函数创建测试数据
- 支持动态数据生成
- 确保测试数据的独立性

### 4. 断言和验证
- 使用 `testify/assert` 进行断言
- 验证 HTTP 状态码和响应内容
- 支持 JSON 响应结构验证

### 5. 错误处理测试
- 测试各种错误场景
- 验证错误响应格式
- 确保错误处理的健壮性

## 使用说明

1. **生成 Mock 文件**：
   ```bash
   mockgen -source=your_handler_file.go -destination=mock_your_repo.go -package=your_package
   ```

2. **替换接口和类型**：
   - 将 `FastAnswerHandler` 替换为你的处理器
   - 将 `FastAnswerRepo` 替换为你的仓储接口
   - 调整路由和请求/响应结构

3. **运行测试**：
   ```bash
   go test -v ./your_handler_package
   go test -bench=. ./your_handler_package
   ```

## 最佳实践

### 1. 测试组织
- 使用 `t.Run()` 组织子测试
- 每个测试用例独立且可重复
- 使用描述性的测试名称

### 2. Mock 使用
- 只 mock 必要的依赖
- 设置明确的期望和验证
- 避免过度 mock

### 3. 测试覆盖
- 测试正常流程和异常流程
- 测试参数验证和错误处理
- 测试边界条件和并发场景

### 4. 性能考虑
- 使用基准测试验证性能
- 避免在测试中执行耗时操作
- 合理使用并发测试

这个模板提供了完整的、生产就绪的 handler 层测试配置，基于 [uber-go/mock](https://github.com/uber-go/mock) 和 Gin 框架，可以直接用于你的项目中。
