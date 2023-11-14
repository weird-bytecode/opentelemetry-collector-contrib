package goldendataset // import "github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/goldendataset"

import (
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
)

func GenerateLogs(logsPairsFile string) ([]plog.Logs, error) {
	pairsData, err := loadPictOutputFile(logsPairsFile)
	if err != nil {
		return nil, err
	}
	pairsTotal := len(pairsData)
	logCombinations := make([]plog.Logs, 0, pairsTotal)
	for _, values := range pairsData {
		logsConfig := &PICTLogInputs{
			ResourceCount:  PICTResourceCount(values[LogResourceCountColumnIndex]),
			ScopeCount:     PICTScopeCount(values[LogScopeCountColoumnIndex]),
			RecordCount:    PICTRecordCount(values[LogRecordCountColumnIndex]),
			RecordSeverity: PICTRecordSeverity(values[LogRecordSeverityColumnIndex]),
		}
		logs := plog.NewLogs()
		appendResourceLogs(*logsConfig, logs)
		logCombinations = append(logCombinations, logs)
	}
	return logCombinations, nil
}

func appendResourceLogs(logConfig PICTLogInputs, logs plog.Logs) {
	switch logConfig.ResourceCount {
	case OneResource:
		resourceLogs := logs.ResourceLogs().AppendEmpty()
		appendLogScopes(logConfig, resourceLogs)
	}
}

func appendLogScopes(logConfig PICTLogInputs, resourceLogs plog.ResourceLogs) {
	switch logConfig.ScopeCount {
	case OneScope:
		scopeLogs := resourceLogs.ScopeLogs().AppendEmpty()
		appendLogRecords(logConfig, scopeLogs)
	}
}

func appendLogRecords(logConfig PICTLogInputs, scopeLogs plog.ScopeLogs) {
	switch logConfig.RecordCount {
	case OneRecord:
		logRecords := scopeLogs.LogRecords()
		logRecords.EnsureCapacity(1)
		record := logRecords.AppendEmpty()

		setLogRecordSeverity(logConfig, record)

		record.Body().SetStr("log record")
		now := pcommon.NewTimestampFromTime(time.Now())
		record.SetTimestamp(now)
	}
}

func setLogRecordSeverity(logConfig PICTLogInputs, record plog.LogRecord) {
	switch logConfig.RecordSeverity {
	case TraceSeverityLevel:
		record.SetSeverityNumber(plog.SeverityNumberTrace)
		record.SetSeverityText("TRACE")
	case DebugSeverityLevel:
		record.SetSeverityNumber(plog.SeverityNumberDebug)
		record.SetSeverityText("DEBUG")
	case InfoSeverityLevel:
		record.SetSeverityNumber(plog.SeverityNumberInfo)
		record.SetSeverityText("INFO")
	case WarnSeverityLevel:
		record.SetSeverityNumber(plog.SeverityNumberWarn)
		record.SetSeverityText("WARN")
	case ErrorSeverityLevel:
		record.SetSeverityNumber(plog.SeverityNumberError)
		record.SetSeverityText("TRACE")
	case FatalSeverityLevel:
		record.SetSeverityNumber(plog.SeverityNumberFatal)
		record.SetSeverityText("TRACE")
	case UnspecifiedSeverityLevel:
	default:
		record.SetSeverityNumber(plog.SeverityNumberUnspecified)
	}
}

const (
	LogResourceCountColumnIndex  = 0
	LogScopeCountColoumnIndex    = 1
	LogRecordCountColumnIndex    = 2
	LogRecordSeverityColumnIndex = 3
)

type PICTLogInputs struct {
	ResourceCount  PICTResourceCount
	ScopeCount     PICTScopeCount
	RecordCount    PICTRecordCount
	RecordSeverity PICTRecordSeverity
}

type PICTRecordSeverity string

// Severity: Unspecified, Trace, Debug, Info, Warn, Error, Fatal

const (
	UnspecifiedSeverityLevel PICTRecordSeverity = "Unspecified"
	TraceSeverityLevel       PICTRecordSeverity = "Trace"
	DebugSeverityLevel       PICTRecordSeverity = "Debug"
	InfoSeverityLevel        PICTRecordSeverity = "Info"
	WarnSeverityLevel        PICTRecordSeverity = "Warn"
	ErrorSeverityLevel       PICTRecordSeverity = "Error"
	FatalSeverityLevel       PICTRecordSeverity = "Fatal"
)

type PICTResourceCount string

const (
	OneResource PICTResourceCount = "One"
)

type PICTScopeCount string

const (
	OneScope PICTScopeCount = "One"
)

type PICTRecordCount string

const (
	OneRecord PICTRecordCount = "One"
)
