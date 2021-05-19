package client

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"sync"
)

var useSSL bool

var gMtx = sync.Mutex{} // Mutex used to sync the gui (mainly for writing to the textview)
var gCtx *guiCtx        // Gui context used by the client (to avoid passing it around everywhere)

// Gui context used by the client
type guiCtx struct {
	chatlog, users *tview.TextView
	pages          *tview.Pages
	app            *tview.Application
}

// Writes text to text log
func writeText(text string) {
	gMtx.Lock()
	gCtx.chatlog.Write([]byte(text))
	gMtx.Unlock()
}

func (c *Client) createGUI() *tview.Application {
	// Sets the style of the GUI
	tview.Styles.SecondaryTextColor = tcell.ColorOrangeRed.TrueColor()
	tview.Styles.ContrastBackgroundColor = tcell.ColorPeachPuff.TrueColor()
	tview.Styles.PrimaryTextColor = tcell.ColorDarkRed.TrueColor()

	// Creates the main app object and its pages
	app = tview.NewApplication()
	pages := tview.NewPages()

	// The chat history page
	users := tview.NewTextView().SetTextAlign(tview.AlignCenter).SetText("USERS")
	text := tview.NewTextView()
	text.SetDynamicColors(true)
	input := tview.NewInputField()
	input.SetFieldBackgroundColor(tcell.ColorBlack)

	grid := tview.NewGrid().
		SetRows(9, 0, 1).
		SetColumns(0, -4, 0).
		SetBorders(true)

	grid.AddItem(users, 0, 0, 3, 1, 0, 100, false).
		AddItem(text, 0, 1, 2, 2, 0, 100, false).
		AddItem(input, 2, 1, 1, 2, 0, 100, true)

	// Bad connection modal
	badConnectionModal := tview.NewModal().
		SetText("Connection could not be made with the server").
		AddButtons([]string{"Ok"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.SwitchToPage("login")
		})

	// Disconnected modal
	disconnectedModal := tview.NewModal().
		SetText("Disconnected from server").
		AddButtons([]string{"Ok"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.SwitchToPage("login")
		})

	// Request failed modal
	requestFailedModal := tview.NewModal().
		SetText("Request failed").
		AddButtons([]string{"Ok"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.SwitchToPage("login")
		})

	// Request successful
	requestSuccededModal := tview.NewModal().
		SetText("Request succeded").
		AddButtons([]string{"Ok"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.SwitchToPage("login")
		})

	// Load the config
	var err error
	details, err = openConfig()
	if err != nil {
		Log.Println("Failed to open config:", err)
	}
	if details == nil {
		details = &Config{}
	}

	// The login page
	var address string
	var username, newname string
	var password string
	var sK, aK string

	form := tview.NewForm().
		AddInputField("Username:", "", 20, nil, func(text string) { username = text; details.Username = text }).
		AddInputField("Password:", "", 20, nil, func(text string) { password = text; details.Password = password }).
		AddInputField("Address:", "", 20, nil, func(text string) { address = text; details.Address = text; c.changeAddress(address) }).
		AddInputField("Server Key:", "", 20, nil, func(text string) { sK = text; details.ServerKey = sK }).
		AddCheckbox("Use SSL:", true, func(checked bool) {
			useSSL = checked
			if checked {
				details.UseSSL = "true"
			} else {
				details.UseSSL = "false"
			}
			c.changeAddress(address)
		})
	form.GetFormItemByLabel("Server Key:").(*tview.InputField).SetMaskCharacter('*')
	form.GetFormItemByLabel("Password:").(*tview.InputField).SetMaskCharacter('*')
	form.SetBorder(true)

	// Loads up the details from the config.json file
	var inputField *tview.InputField
	inputField = form.GetFormItemByLabel("Username:").(*tview.InputField)
	if details.Username != "" {
		inputField.SetText(details.Username)
	}
	inputField = form.GetFormItemByLabel("Password:").(*tview.InputField)
	if details.Password != "" {
		inputField.SetText(details.Password)
	}
	inputField = form.GetFormItemByLabel("Address:").(*tview.InputField)
	if details.Address != "" {
		inputField.SetText(details.Address)
	}
	inputField = form.GetFormItemByLabel("Server Key:").(*tview.InputField)
	if details.ServerKey != "" {
		inputField.SetText(details.ServerKey)
	}
	inputCheckbox := form.GetFormItemByLabel("Use SSL:").(*tview.Checkbox)
	if details.UseSSL == "false" {
		inputCheckbox.SetChecked(false)
	}

	// Button for connecting the user to the server
	form.AddButton("Connect", func() {
		// Save the config when connecting
		err := saveConfig(details)
		if err != nil {
			Log.Println("Failed to save config:", err)
		}

		// Attempt to connect
		err = c.connect()
		if err != nil {
			// Show it failed on the GUI
			Log.Println("Failed to connect:", err)
			pages.SwitchToPage("badConnection")
		} else {
			// On success the user is moved to the chat screen
			Log.Println("Successful connection")
			pages.SwitchToPage("spotify")

			// Binds the input bar to send messages to the server on "Enter" keypress
			input.SetDoneFunc(func(key tcell.Key) {
				if key == tcell.KeyEnter {
					writeText("[#343434]>> " + input.GetText() + "\n")
					c.writeMsg(input.GetText())
					input.SetText("")
				}
			})

			// Creates the gui context used by the client
			gCtx = &guiCtx{
				chatlog: text,
				users:   users,
				app:     app,
				pages:   pages,
			}

			// Listen for incoming messages
			go c.readPump()

			// Handle shutdown
			go c.handleShutdown()
		}
	})

	// Button for creating an account in the server
	form.AddButton("Create Account", func() {
		var err error

		err = c.createUser(username, password, sK, aK)
		if err != nil {
			pages.SwitchToPage("requestFailed")
		} else {
			pages.SwitchToPage("requestSucceded")
		}
	})

	// Button for creating an account in the server
	form.AddButton("Admin", func() {
		pages.SwitchToPage("admin")
	})

	// Button for quitting the app
	form.AddButton("Quit", func() {
		// Save the config when quitting
		err := saveConfig(details)
		if err != nil {
			Log.Println("Failed to save config:", err)
		}

		app.Stop()
	})

	// Screen for admin functionality
	adminForm := tview.NewForm().
		AddInputField("Current Name:", "", 20, nil, func(text string) { username = text; details.Username = text }).
		AddInputField("New Password (if applicable):", "", 20, nil, func(text string) { password = text; details.Password = password }).
		AddInputField("New Name (if applicable):", "", 20, nil, func(text string) { newname = text }).
		AddInputField("Admin Key:", "", 20, nil, func(text string) { aK = text; details.AdminKey = aK }).
		AddInputField("Address:", "", 20, nil, func(text string) { address = text; details.Address = text; c.changeAddress(address) })
	adminForm.GetFormItemByLabel("Admin Key:").(*tview.InputField).SetMaskCharacter('*')
	adminForm.GetFormItemByLabel("New Password (if applicable):").(*tview.InputField).SetMaskCharacter('*')
	adminForm.SetBorder(true)

	inputField = adminForm.GetFormItemByLabel("Current Name:").(*tview.InputField)
	if details.Username != "" {
		inputField.SetText(details.Username)
	}
	inputField = adminForm.GetFormItemByLabel("Address:").(*tview.InputField)
	if details.Address != "" {
		inputField.SetText(details.Address)
	}
	inputField = adminForm.GetFormItemByLabel("Admin Key:").(*tview.InputField)
	if details.AdminKey != "" {
		inputField.SetText(details.AdminKey)
	}

	// Button for going back to the login page
	adminForm.AddButton("Go Back", func() {
		pages.SwitchToPage("login")
	})

	// Button for updating account details
	adminForm.AddButton("Update Account", func() {
		var err error

		err = c.updateUser(username, newname, password, aK)
		if err != nil {
			pages.SwitchToPage("requestFailed")
		} else {
			pages.SwitchToPage("requestSucceded")
		}
	})

	// Button for deleting an account in the server
	adminForm.AddButton("Delete Account", func() {
		var err error

		err = c.deleteUser(username, aK)
		if err != nil {
			pages.SwitchToPage("requestFailed")
		} else {
			pages.SwitchToPage("requestSucceded")
		}
	})

	// Add each page to the pages object to enable switching to different screens
	pages.AddPage("login", form, true, true)
	pages.AddPage("admin", adminForm, true, false)
	pages.AddPage("spotify", grid, true, false)
	pages.AddPage("badConnection", badConnectionModal, true, false)
	pages.AddPage("requestFailed", requestFailedModal, true, false)
	pages.AddPage("requestSucceded", requestSuccededModal, true, false)
	pages.AddPage("disconnected", disconnectedModal, true, false)
	app.SetRoot(pages, true).EnableMouse(true)

	return app
}
