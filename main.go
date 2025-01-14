//go:generate fyne bundle -o bundled.go Icon.png

package main

import (
	"flag"
	"image/color"
	"log"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"golang.org/x/crypto/ssh"

	"github.com/fyne-io/terminal"
	"github.com/fyne-io/terminal/cmd/fyneterm/data"
)

type termResizer struct {
	widget.Icon

	term  *terminal.Terminal
	debug bool
	sess  *ssh.Session
	win   fyne.Window
}

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "Show terminal debug messages")
	flag.Parse()

	a := app.NewWithID("xyz.andy.sshterm2")
	a.Settings().SetTheme(newTermTheme())
	w := a.NewWindow("SSH Terminal")
	a.SetIcon(resourceIconPng)
	w.SetPadded(false)

	bg := canvas.NewRectangle(color.Gray{Y: 0x16})
	img := canvas.NewImageFromResource(data.FyneLogo)
	img.FillMode = canvas.ImageFillContain
	img.Translucency = 0.95

	t := &termResizer{win: w, debug: debug}
	t.ExtendBaseWidget(t)
	w.SetContent(container.NewMax(bg, img, t))

	cellSize := guessCellSize()
	w.Resize(fyne.NewSize(cellSize.Width*80, cellSize.Height*24))

	askForSSH(t, w, a)
	w.ShowAndRun()
}

func (r *termResizer) Resize(s fyne.Size) {
	r.Icon.Resize(s)
	if r.sess == nil {
		return
	}
	cellSize := guessCellSize()
	err := r.sess.WindowChange(int(s.Height/cellSize.Height), int(s.Width/cellSize.Width))
	if err != nil {
		log.Println("Failed to resize", err)
	}
}

func askForSSH(t *termResizer, w fyne.Window, a fyne.App) {
	host := widget.NewEntryWithData(binding.BindPreferenceString("login.host", a.Preferences()))
	user := widget.NewEntryWithData(binding.BindPreferenceString("login.user", a.Preferences()))
	pass := widget.NewPasswordEntry()

	d := dialog.NewForm("SSH Connection Details", "Connect", "Clear",
		[]*widget.FormItem{
			widget.NewFormItem("Host", host),
			widget.NewFormItem("Username", user),
			widget.NewFormItem("Password", pass),
		}, func(ok bool) {
			if !ok {
				a.Preferences().SetString("login.host", "")
				a.Preferences().SetString("login.user", "")
				askForSSH(t, w, a)
				return
			}

			runSSH(host.Text, user.Text, pass.Text, t, w, a)
		}, w)
	d.Show()

	host.OnSubmitted = func(_ string) {
		w.Canvas().Focus(user)
	}
	user.OnSubmitted = func(_ string) {
		w.Canvas().Focus(pass)
	}
	pass.OnSubmitted = func(_ string) {
		d.Submit()
	}

	if host.Text == "" {
		w.Canvas().Focus(host)
	} else {
		w.Canvas().Focus(pass)
	}
}

func (r *termResizer) Tapped(_ *fyne.PointEvent) {
	r.win.Canvas().Focus(r.term)
}

func runSSH(host, user, pass string, t *termResizer, w fyne.Window, a fyne.App) {
	if !strings.Contains(host, ":") {
		host = host + ":22"
	}
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(pass),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	conn, err := ssh.Dial("tcp", host, config)
	if err != nil {
		showError(err, t, w, a)
		return
	}

	session, err := conn.NewSession()
	if err != nil {
		showError(err, t, w, a)
		return
	}
	t.sess = session

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	cellSize := guessCellSize()
	if err := session.RequestPty("xterm-256color", int(t.Size().Height/cellSize.Height), int(t.Size().Width/cellSize.Width), modes); err != nil {
		_ = session.Close()
		showError(err, t, w, a)
		return
	}

	in, _ := session.StdinPipe()
	out, _ := session.StdoutPipe()

	go session.Run("$SHELL || bash")

	go func() {
		go func() {
			time.Sleep(100 * time.Millisecond)
			t.Tapped(nil) // focus/mobile keyboard workaround
		}()

		t.term = terminal.New()
		t.term.SetDebug(t.debug)
		c := w.Content().(*fyne.Container)
		w.SetContent(container.NewMax(c.Objects[0], c.Objects[1], t.term, c.Objects[len(c.Objects)-1]))

		_ = t.term.RunWithConnection(in, out)

		t.term = nil
		w.SetContent(container.NewMax(c.Objects[0], c.Objects[1], c.Objects[len(c.Objects)-1]))
		askForSSH(t, w, a)
	}()
}

func guessCellSize() fyne.Size {
	cell := canvas.NewText("M", color.White)
	cell.TextStyle.Monospace = true

	return cell.MinSize()
}

func showError(err error, t *termResizer, w fyne.Window, a fyne.App) {
	d := dialog.NewError(err, w)
	d.SetOnClosed(func() {
		askForSSH(t, w, a)
	})
	d.Show()
}
