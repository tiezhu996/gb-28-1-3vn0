package util

import (
	"bytes"
	"errors"
	"strings"

	"github.com/xuri/excelize/v2"

	"github.com/onlineexam/onlineexam/internal/constants"
)

// ExcelQuestionRow 题目 Excel 导入行结构（模板列固定）。
// 列顺序：题型 | 学科 | 知识点(逗号分隔) | 难度 | 题干 | 选项A | 选项B | 选项C | 选项D | 答案 | 解析 | 分值
type ExcelQuestionRow struct {
	Type           string   `json:"type"`
	Subject        string   `json:"subject"`
	KnowledgePoint string   `json:"knowledge_point"`
	Difficulty     string   `json:"difficulty"`
	Content        string   `json:"content"`
	Options        []string `json:"options"`
	Answer         string   `json:"answer"`
	Analysis       string   `json:"analysis"`
	Score          float64  `json:"score"`
}

// ParseQuestionExcel 解析 Excel 题目模板文件，返回行数据。
func ParseQuestionExcel(fileBytes []byte) ([]ExcelQuestionRow, error) {
	f, err := excelize.OpenReader(bytes.NewReader(fileBytes))
	if err != nil {
		return nil, WrapAppError(constants.CodeQuestionImportErr, "题库模块：Excel 文件解析失败", err)
	}
	defer func() { _ = f.Close() }()

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return nil, WrapAppError(constants.CodeQuestionImportErr, "题库模块：读取 Sheet1 失败", err)
	}
	if len(rows) < 2 {
		return nil, WrapAppError(constants.CodeQuestionImportErr, "题库模块：Excel 模板为空，至少需要一行数据", errors.New("empty excel"))
	}

	var result []ExcelQuestionRow
	for idx, row := range rows[1:] {
		if len(row) == 0 || strings.TrimSpace(row[0]) == "" {
			continue
		}
		r := ExcelQuestionRow{
			Type:           strings.TrimSpace(cell(row, 0)),
			Subject:        strings.TrimSpace(cell(row, 1)),
			KnowledgePoint: strings.TrimSpace(cell(row, 2)),
			Difficulty:     strings.TrimSpace(cell(row, 3)),
			Content:        strings.TrimSpace(cell(row, 4)),
			Options:        []string{strings.TrimSpace(cell(row, 5)), strings.TrimSpace(cell(row, 6)), strings.TrimSpace(cell(row, 7)), strings.TrimSpace(cell(row, 8))},
			Answer:         strings.TrimSpace(cell(row, 9)),
			Analysis:       strings.TrimSpace(cell(row, 10)),
		}
		score := strings.TrimSpace(cell(row, 11))
		if score == "" {
			r.Score = 5
		} else {
			var s float64
			if _, err := fmtSscan(score, &s); err != nil {
				r.Score = 5
			} else {
				r.Score = s
			}
		}
		result = append(result, r)
		_ = idx
	}
	if len(result) == 0 {
		return nil, WrapAppError(constants.CodeQuestionImportErr, "题库模块：未解析到任何题目行", errors.New("no valid rows"))
	}
	return result, nil
}

func cell(row []string, i int) string {
	if i < len(row) {
		return row[i]
	}
	return ""
}
