package logs

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/testbed/correctnesstests"
	"github.com/open-telemetry/opentelemetry-collector-contrib/testbed/testbed"
	"github.com/stretchr/testify/require"
)

var correctnessResults testbed.TestResultsSummary = &testbed.CorrectnessResults{}

func TestMain(m *testing.M) {
	testbed.DoTestMain(m, correctnessResults)
}

func TestLogs(t *testing.T) {
	tests, err := correctnesstests.LoadPictOutputPipelineDefs("testdata/generated_pict_pairs_logs_pipeline.txt")
	require.NoError(t, err)
	processors := map[string]string{
		"batch": `
  batch:
    send_batch_size: 1024
`,
	}
	for _, test := range tests {
		test.TestName = fmt.Sprintf("%s-%s", test.Receiver, test.Exporter)
		test.DataSender = correctnesstests.ConstructLogSender(t, test.Receiver)
		test.DataReceiver = correctnesstests.ConstructReceiver(t, test.Exporter)
		t.Run(test.TestName, func(t *testing.T) {
			testWithLogsGoldenDataset(t, test.DataSender, test.DataReceiver, test.ResourceSpec, processors)
		})
	}
}

func testWithLogsGoldenDataset(
	t *testing.T,
	sender testbed.DataSender,
	receiver testbed.DataReceiver,
	resourceSpec testbed.ResourceSpec,
	processors map[string]string,
) {
	dataProvider := testbed.NewGoldenDataProvider(
		"",
		"",
		"",
		"../../../internal/coreinternal/goldendataset/testdata/generated_pict_pairs_logs.txt")
	factories, err := testbed.Components()
	require.NoError(t, err, "default components resulted in: %v", err)
	runner := testbed.NewInProcessCollector(factories)
	validator := testbed.NewCorrectTestValidator(sender.ProtocolName(), receiver.ProtocolName(), dataProvider)
	config := correctnesstests.CreateConfigYaml(sender, receiver, processors, "logs")
	log.Println(config)
	configCleanup, cfgErr := runner.PrepareConfig(config)
	require.NoError(t, cfgErr, "collector configuration resulted in: %v", cfgErr)
	defer configCleanup()
	tc := testbed.NewTestCase(
		t,
		dataProvider,
		sender,
		receiver,
		runner,
		validator,
		correctnessResults,
		testbed.WithResourceLimits(resourceSpec),
	)
	defer tc.Stop()

	tc.EnableRecording()
	tc.StartBackend()
	tc.StartAgent()

	tc.StartLoad(testbed.LoadOptions{
		DataItemsPerSecond: 1024,
		ItemsPerBatch:      1,
	})

	tc.Sleep(2 * time.Second)

	tc.StopLoad()

	tc.WaitForN(func() bool { return tc.LoadGenerator.DataItemsSent() == tc.MockBackend.DataItemsReceived() },
		3*time.Second, "all data items received")

	tc.StopAgent()

	tc.ValidateData()
}
