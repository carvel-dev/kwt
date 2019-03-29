package workspace

import (
	"k8s.io/client-go/rest"
)

type WorkspaceDesktop struct {
	Workspace Workspace
}

func (w WorkspaceDesktop) Install(restConfig *rest.Config) error {
	// TODO detect resolution system_profiler SPDisplaysDataType|grep Resolution
	execOpts := ExecuteOpts{
		Command: []string{"/bin/bash"},
		CommandArgs: []string{"-c", `
			set -e -x

			apt-get -y update
			apt-get -y install x11vnc xvfb openbox
			
			export DISPLAY=:1
			nohup Xvfb :1 -screen 0 1024x768x16 </dev/null >/tmp/xvfb.log 2>&1 &
			sleep 5
			nohup x11vnc -display :1 -bg -listen 0.0.0.0 -xkb -passwd 123456 -noxrecord -noxfixes -noxdamage -forever </dev/null >/tmp/vnc.log 2>&1
			nohup openbox </dev/null >/tmp/openbox.log 2>&1 &
		`},
	}

	return w.Workspace.Execute(execOpts, restConfig)
}

func (w WorkspaceDesktop) AddFirefox(restConfig *rest.Config) error {
	// TODO // https://www.ghacks.net/2016/07/22/multi-process-firefox/ ?
	execOpts := ExecuteOpts{
		Command: []string{"/bin/bash"},
		CommandArgs: []string{"-c", `
			set -e -x

			apt-get update
			apt-get -y install firefox
		`},
	}

	return w.Workspace.Execute(execOpts, restConfig)
}

func (w WorkspaceDesktop) AddSublimeText(restConfig *rest.Config) error {
	execOpts := ExecuteOpts{
		Command: []string{"/bin/bash"},
		CommandArgs: []string{"-c", `
			set -e -x
			
			apt-get update
			apt-get -y install sudo wget
			
			wget -qO - https://download.sublimetext.com/sublimehq-pub.gpg | sudo apt-key add -
			sudo apt-get -y install apt-transport-https
			echo "deb https://download.sublimetext.com/ apt/stable/" | sudo tee /etc/apt/sources.list.d/sublime-text.list
			
			sudo apt-get update
			sudo apt-get -y install sublime-text libgtk2.0-0
		`},
	}

	return w.Workspace.Execute(execOpts, restConfig)
}

func (w WorkspaceDesktop) AddChrome(restConfig *rest.Config) error {
	execOpts := ExecuteOpts{
		Command: []string{"/bin/bash"},
		CommandArgs: []string{"-c", `
			set -e -x
			
			apt-get update
			apt-get -y install sudo wget

			wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | sudo apt-key add -
			echo 'deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main' | sudo tee /etc/apt/sources.list.d/google-chrome.list
			
			sudo apt-get update
			sudo apt-get -y install google-chrome-stable
		`},
	}

	return w.Workspace.Execute(execOpts, restConfig)
}

func (w WorkspaceDesktop) AddGo1x(restConfig *rest.Config) error {
	execOpts := ExecuteOpts{
		Command: []string{"/bin/bash"},
		CommandArgs: []string{"-c", `
			set -e -x

			apt-get update
			apt-get -y install sudo vim wget

			wget -q -O - https://dl.google.com/go/go1.11.4.linux-amd64.tar.gz > /tmp/go_1_11.tgz
			tar xzvf /tmp/go_1_11.tgz -C /usr/local

			echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
		`},
	}

	return w.Workspace.Execute(execOpts, restConfig)
}

func (w WorkspaceDesktop) AddDocker(restConfig *rest.Config) error {
	execOpts := ExecuteOpts{
		Command: []string{"/bin/bash"},
		CommandArgs: []string{"-c", `
			set -e -x

			apt-get update
			apt-get -y install apt-transport-https ca-certificates curl gnupg-agent software-properties-common

			curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -

			add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
			apt-get update

			apt-get install -y docker-ce docker-ce-cli containerd.io
		`},
	}

	return w.Workspace.Execute(execOpts, restConfig)
}
