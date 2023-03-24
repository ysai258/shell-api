package main

import (
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
)

// declaring format of request json body
type CmdRequest struct {
	Command string `json:"command" binding:"required"`
}

func main() {
	r := gin.Default()

	/*
		 	NOTE Inside commandHandler change the shell you have inside the PATH
			for example : my current mac uses zsh it may vary to your system maybe bash
			so change that according , default using shell
	*/

	r.POST("/api/cmd", commandHandler)

	if err := r.Run(); err != nil {
		log.Fatal("unable to start server: ", err.Error())
	}
}

func commandHandler(c *gin.Context) {
	var cmd CmdRequest

	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	if !IsSafeCommand(cmd.Command) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsafe command"})
		return
	}
	// based on deployed runtime os changing shell
	shell := ""
	shellArg := ""

	switch runtime.GOOS { // getting current runtime os
	case "darwin":
		shell = "zsh"
		shellArg = "-c"

	case "windows":
		shell = "cmd"
		shellArg = "/C"

	default:
		shell = "shell"
		shellArg = "-c"
	}
	out, err := exec.Command(shell, shellArg, cmd.Command).CombinedOutput()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"output": string(out), "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"output": string(out)})
}

func IsSafeCommand(cmd string) bool {
	forbiddenCommands := map[string]bool{
		"rm":       true,
		"sudo":     true,
		"reboot":   true,
		"halt":     true,
		"shutdown": true,
		"poweroff": true,
		"mkfs":     true,
		"dd":       true,
	}

	parts := strings.Split(cmd, " ")

	for _, part := range parts {
		if forbiddenCommands[part] {
			return false
		}
	}

	return true
}
