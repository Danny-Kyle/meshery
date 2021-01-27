package system

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/layer5io/meshery/mesheryctl/internal/cli/root/config"
	"github.com/layer5io/meshery/mesheryctl/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var b *bytes.Buffer

func SetupContextEnv(t *testing.T) {
	path, err := os.Getwd()
	if err != nil {
		t.Error("unable to locate meshery directory")
	}
	viper.Reset()
	viper.SetConfigFile(path + "/../../../../pkg/utils/TestConfig.yaml")
	//fmt.Println(viper.ConfigFileUsed())
	err = viper.ReadInConfig()
	if err != nil {
		t.Errorf("unable to read configuration from %v, %v", viper.ConfigFileUsed(), err.Error())
	}

	mctlCfg, err = config.GetMesheryCtl(viper.GetViper())
	if err != nil {
		t.Error("error processing config", err)
	}
}

func SetupFunc(t *testing.T) {
	//fmt.Println(viper.AllKeys())
	b = bytes.NewBufferString("")
	logrus.SetOutput(b)
	logrus.SetFormatter(&utils.OnlyStringFormatterForLogrus{})
	SystemCmd.SetOut(b)
}

func BreakupFunc(t *testing.T) {
	viewCmd.Flags().VisitAll(setFlagValueAsUndefined)
}

func setFlagValueAsUndefined(flag *pflag.Flag) {
	_ = flag.Value.Set("")
}

type CmdTestInput struct {
	Name             string
	Args             []string
	ExpectedResponse string
}

func TestViewCmd(t *testing.T) {
	SetupContextEnv(t)
	expectedResponseForAll := ""
	for k, v := range mctlCfg.Contexts {
		expectedResponseForAll += PrintChannelAndVersionToStdout(v, k) + "\n\n"
	}
	expectedResponseForAll += fmt.Sprintf("Current Context: %v\n", mctlCfg.CurrentContext)

	tests := []CmdTestInput{
		{
			Name:             "view without any parameter",
			Args:             []string{"channel", "view"},
			ExpectedResponse: PrintChannelAndVersionToStdout(mctlCfg.Contexts["local"], "local") + "\n\n",
		},
		{
			Name:             "view with context override",
			Args:             []string{"channel", "view", "-c", "gke"},
			ExpectedResponse: PrintChannelAndVersionToStdout(mctlCfg.Contexts["gke"], "gke") + "\n\n",
		},
		{
			Name:             "view with all flag",
			Args:             []string{"channel", "view", "--all"},
			ExpectedResponse: expectedResponseForAll,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			SetupFunc(t)
			SystemCmd.SetArgs(tt.Args)
			err = SystemCmd.Execute()
			if err != nil {
				t.Error(err)
			}

			actualResponse := b.String()
			expectedResponse := tt.ExpectedResponse

			if expectedResponse != actualResponse {
				t.Errorf("expected response %v and actual response %v don't match", expectedResponse, actualResponse)
			}
			BreakupFunc(t)
		})
	}
}
func TestRunChannelWithNoCmdOrFlag(t *testing.T) {
	SetupFunc(t)
	SystemCmd.SetArgs([]string{"channel"})
	err = SystemCmd.Execute()
	if err != nil {
		t.Error(err)
	}

	actualResponse := b.String()
	expectedResponse := ""
	expectedResponse += PrintChannelAndVersionToStdout(mctlCfg.GetContextContent(), "local") + "\n\n"
	expectedResponse += channelCmd.UsageString()

	if expectedResponse != actualResponse {
		t.Errorf("expected response %v and actual response %v don't match", expectedResponse, actualResponse)
	}
	BreakupFunc(t)
}
