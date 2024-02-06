// Implementations for the handler routines when the payload is being built
package builder

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	thanatoserror "thanatos/errors"

	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"github.com/MythicMeta/MythicContainer/mythicrpc"
)

const AGENT_CODE_PATH = "../agent"

// Type for the handler routines when being built by Mythic
type MythicPayloadHandler struct{}

// This will build the agent using the specified command string
func (handler MythicPayloadHandler) Build(target string, config ParsedPayloadParameters, command string) ([]byte, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return []byte{}, thanatoserror.Errorf("failed to get the current working directory: %s", err.Error())
	}

	agentCodePath := fmt.Sprintf("%s/%s", cwd, AGENT_CODE_PATH)

	cmd := exec.Command("/bin/bash", "-c", command)
	cmd.Dir = agentCodePath

	output, err := cmd.CombinedOutput()
	if err != nil {
		errorMsg := thanatoserror.Errorf("failed to build the agent: %s", err.Error())
		return []byte{}, errors.Join(thanatoserror.Errorf("output for command '/bin/bash -c %s:\n%s", command, string(output)), errorMsg)
	}

	outpath := fmt.Sprintf("%s/target/%s/release", agentCodePath, target)

	profile := ""
	if config.C2Profiles.HttpProfile != nil {
		profile = "http"
	} else {
		panic("Unimplemented build profile")
	}

	filename := ""
	if config.SelectedOS == agentstructs.SUPPORTED_OS_LINUX {
		if config.PayloadBuildParameters.Output != PayloadBuildParameterOutputFormatExecutable {
			filename = fmt.Sprintf("libthanatos_%s_cdylib.so", profile)
		} else {
			filename = fmt.Sprintf("thanatos_%s_binary", profile)
		}
	} else if config.SelectedOS == agentstructs.SUPPORTED_OS_WINDOWS {
		if config.PayloadBuildParameters.Output == PayloadBuildParameterOutputFormatExecutable {
			filename = fmt.Sprintf("thanatos_%s_binary.exe", profile)
		} else {
			filename = fmt.Sprintf("thanatos_%s_cdylib.dll", profile)
		}
	}

	payload, err := os.ReadFile(fmt.Sprintf("%s/%s", outpath, filename))
	if err != nil {
		return []byte{}, thanatoserror.Errorf("failed to open the built payload: %s", err.Error())
	}

	return payload, nil
}

// This updates the current build step in Mythic
func (handler MythicPayloadHandler) UpdateBuildStep(input mythicrpc.MythicRPCPayloadUpdateBuildStepMessage) (*mythicrpc.MythicRPCPayloadUpdateBuildStepMessageResponse, error) {
	return mythicrpc.SendMythicRPCPayloadUpdateBuildStep(input)
}