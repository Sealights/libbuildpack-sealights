package sealights

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/libbuildpack"
)

const AgentName = "SL.DotNet.dll"
const DefaultAgentMode = "testListener"

type Launcher struct {
	Log       *libbuildpack.Logger
	Options   *SealightsOptions
	AgentDir  string
	DotNetDir string
}

func NewLauncher(log *libbuildpack.Logger, options *SealightsOptions, agentInstallationDir string, dotnetInstallationDir string) *Launcher {
	return &Launcher{Log: log, Options: options, AgentDir: agentInstallationDir, DotNetDir: dotnetInstallationDir}
}

func (la *Launcher) ModifyStartParameters(stager *libbuildpack.Stager) error {
	la.updateAgentPath(stager)

	releaseInfo := NewReleaseInfo(stager.BuildDir())

	startCommand := releaseInfo.GetStartCommand()
	newStartCommand := la.updateStartCommand(startCommand)
	err := releaseInfo.SetStartCommand(newStartCommand)
	if err != nil {
		return err
	}

	la.Log.Info(fmt.Sprintf("Sealights: Start command updated. From '%s' to '%s'", startCommand, newStartCommand))

	return nil
}

func (la *Launcher) updateAgentPath(stager *libbuildpack.Stager) {
	if strings.HasPrefix(la.AgentDir, stager.BuildDir()) {
		clearPath := strings.TrimPrefix(la.AgentDir, stager.BuildDir())
		la.AgentDir = filepath.Join(".", clearPath)
	}
}

func (la *Launcher) updateStartCommand(originalCommand string) string {
	// expected command format:
	// cd ${DEPS_DIR}/0/dotnet_publish && exec ./app --server.urls http://0.0.0.0:${PORT}
	// cd ${DEPS_DIR}/0/dotnet_publish && exec dotnet ./app.dll --server.urls http://0.0.0.0:${PORT}

	parts := strings.SplitAfterN(originalCommand, "exec ", 2)

	newCmd := parts[0] + la.buildCommandLine(parts[1])

	return newCmd
}

//dotnet SL.DotNet.dll testListener --logAppendFile true --logFilename /tmp/collector.log --tokenFile /tmp/sltoken.txt --buildSessionIdFile /tmp/buildsessionid.txt --target dotnet --workingDir /tmp/app --profilerLogDir /tmp/ --profilerLogLevel 7 --targetArgs \"test app.dll\"
func (la *Launcher) buildCommandLine(command string) string {
	if la.Options.CustomCommand != "" {
		return la.Options.CustomCommand
	}

	var sb strings.Builder
	options := la.Options

	agent := filepath.Join(la.AgentDir, AgentName)
	dotnetCli := "dotnet"
	if la.DotNetDir != "" {
		dotnetCli = filepath.Join(la.DotNetDir, "dotnet")
	}

	agentMode := DefaultAgentMode
	if options.Verb != "" {
		agentMode = options.Verb
	}

	sb.WriteString(fmt.Sprintf("%s %s %s", dotnetCli, agent, agentMode))

	for key, value := range la.Options.SlArguments {
		sb.WriteString(fmt.Sprintf(" --%s %s", key, value))
	}

	if la.Options.ParseArgsFromCmd == "true" {
		_, exists := la.Options.SlArguments["workingDir"]
		if (!exists){
			sb.WriteString(" --workingDir ${PWD}")
		}

		parsedTarget, parsedArgs := la.getTargetArgs(command)
		_, exists = la.Options.SlArguments["target"]
		if (!exists){
			sb.WriteString(fmt.Sprintf(" --target %s", parsedTarget))
		}

		_, exists = la.Options.SlArguments["targetArgs"]
		if (!exists){
			sb.WriteString(fmt.Sprintf(" --targetArgs \"%s\"", parsedArgs))
		}
	}

	return sb.String()
}

func (la *Launcher) getTargetArgs(command string) (target string, args string) {
	if strings.HasPrefix(command, "dotnet") || la.DotNetDir == "" {
		// use dotnet as target and remove it from command
		target = "dotnet"
		command = strings.TrimPrefix(command, "dotnet")
		command = strings.TrimPrefix(command, " ")
	} else {
		// use dotnet from sealights folder
		target = filepath.Join(la.DotNetDir, "dotnet")
	}

	parts := strings.SplitN(command, " ", 2)
	withoutArguments := parts[0]
	args = fmt.Sprintf("test %s", withoutArguments)

	if strings.HasPrefix(args, "--") {
		args = fmt.Sprintf(" %s", args)
	}

	return
}
