package loggingformatting

import (
	"go/ast"

	"github.com/retroenv/retrogolint/internal/rules/api"
)

func extractSprintfFormatFromLogStringField(fieldCall *ast.CallExpr) (string, bool) {
	if !api.IsLogStringCall(fieldCall) || len(fieldCall.Args) < 2 {
		return "", false
	}

	sprintfCall, ok := fieldCall.Args[1].(*ast.CallExpr)
	if !ok || !api.IsFmtSprintfCall(sprintfCall) || len(sprintfCall.Args) < 1 {
		return "", false
	}

	return api.ExtractStringLiteral(sprintfCall.Args[0])
}
