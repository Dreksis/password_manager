// Import necessary packages
package main

import (
    "fmt"
    "os"
    "strings"
    "log"
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/theme"
    "fyne.io/fyne/v2/widget"
    "golang.org/x/crypto/bcrypt"
    "encoding/json"
    "io/ioutil"
)

// Define a struct to hold service passwords
type Passwords struct {
    ServicePasswords map[string]string
}

// Define a struct to manage passwords
type PasswordManager struct {
    passwords map[string]string
}

// Create a new password manager
func NewPasswordManager() *PasswordManager {
    return &PasswordManager{
        passwords: make(map[string]string),
    }
}

// Add a new password to the password manager
func (pm *PasswordManager) AddPassword(service, password string) error {
    // Check if password for service already exists
    if _, ok := pm.passwords[service]; ok {
        return fmt.Errorf("password for service %s already exists", service)
    }

    // Hash the password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return fmt.Errorf("failed to hash password: %v", err)
    }

    // Add the hashed password to the password manager
    pm.passwords[service] = string(hashedPassword)

    return nil
}

// Get a password from the password manager
func (pm *PasswordManager) GetPassword(service, password string) (string, error) {
    // Get the stored hash for the service
    storedHash, exists := pm.passwords[service]
    if !exists {
        return "", fmt.Errorf("No password found for service: %s", service)
    }

    // Compare the entered password to the stored hash
    err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
    if err != nil {
        return "", fmt.Errorf("Incorrect password for service: %s", service)
    }

    return storedHash, nil
}

// Define a container to hold the content of the application
var content = container.NewVBox()

// Synchronize the password manager with the passwords file
func synchronizeData(pm *PasswordManager, passwords *Passwords) {
    for service, password := range pm.passwords {
        passwords.ServicePasswords[service] = password
    }
}

// Create a new Passwords struct
func NewPasswords() *Passwords {
    return &Passwords{
        ServicePasswords: make(map[string]string),
    }
}

// Load passwords from a file
func loadPasswordsFromFile(filepath string) (*Passwords, error) {
    // Read the file
    data, err := os.ReadFile(filepath)
    if err != nil && os.IsNotExist(err) {
        // If the file doesn't exist, create a new file with an empty Passwords struct
        log.Printf("File does not exist, initializing new data: %v", err)
        emptyPasswords := NewPasswords()
        err = savePasswordsToFile(emptyPasswords, filepath)
        if err != nil {
            log.Printf("Error creating initial JSON: %v", err)
            return nil, err
        }
        return emptyPasswords, nil
    } else if err != nil {
        // Handle other potential errors from os.ReadFile
        log.Printf("Error reading file: %v", err)
        return nil, err
    }

    // Unmarshal the JSON data into a Passwords struct
    var passwords Passwords
    err = json.Unmarshal(data, &passwords.ServicePasswords)
    if err != nil {
        log.Printf("Error unmarshaling JSON: %v", err)
        return &Passwords{}, err
    }

    return &passwords, nil
}

// Save passwords to a file
func savePasswordsToFile(passwords *Passwords, filepath string) error {
    if passwords == nil {
        return fmt.Errorf("passwords data is nil")
    }

    // Serialize the Passwords struct to JSON
    data, err := json.Marshal(passwords.ServicePasswords)
    if err != nil {
        log.Printf("Error serializing data: %v", err)
        return err
    }

    // Write the JSON data to the file
    err = ioutil.WriteFile(filepath, data, 0644)
    if err != nil {
        log.Printf("Error writing to file: %v", err)
        return err
    }

    return nil
}

// Main function
func main() {
    // Create a new password manager
    pm := NewPasswordManager()

    // Create a new application
    a := app.New()

    // Create a new window
    w := a.NewWindow("Password Manager")

    // Create entry fields for adding a password
    serviceEntry := widget.NewEntry()
    serviceEntry.SetPlaceHolder("Service")
    passwordEntry := widget.NewPasswordEntry()
    passwordEntry.SetPlaceHolder("Password")

    // Set the theme to DarkTheme
    a.Settings().SetTheme(theme.DarkTheme())

    // Create a button to add a password
    addButton := widget.NewButton("Add", func() {
        service := strings.TrimSpace(serviceEntry.Text)
        password := strings.TrimSpace(passwordEntry.Text)
    
        errMsg := pm.AddPassword(service, password)
        if errMsg != nil {
            errLabel := widget.NewLabel(errMsg.Error()) // Convert error to string
            content.Add(errLabel)
            return
        }
    
        serviceEntry.SetText("")
        passwordEntry.SetText("")
    })

    // Create entry fields for getting a password
    getServiceEntry := widget.NewEntry()
    getServiceEntry.SetPlaceHolder("Service")
    getPasswordEntry := widget.NewPasswordEntry()
    getPasswordEntry.SetPlaceHolder("Password")

    // Create a button to get a password
    getButton := widget.NewButton("Get", func() {
        service := strings.TrimSpace(getServiceEntry.Text)
        password := strings.TrimSpace(getPasswordEntry.Text)
    
        hashedPassword, err := pm.GetPassword(service, password)
        if err != nil {
            errLabel := widget.NewLabel(err.Error()) // Convert error to string
            content.Add(errLabel)
            return
        }
    
        passwordLabel := widget.NewLabel(fmt.Sprintf("Password for service %s: %s", service, hashedPassword))
        content.Add(passwordLabel)
    })

    // Create a button to quit the application
    quitButton := widget.NewButton("Quit", func() {
        os.Exit(0)
    })

    // Create a form for adding a password
    addForm := widget.NewForm(
        &widget.FormItem{Text: "Service", Widget: serviceEntry},
        &widget.FormItem{Text: "Password", Widget: passwordEntry},
    )

    // Create a form for getting a password
    getForm := widget.NewForm(
        &widget.FormItem{Text: "Service", Widget: getServiceEntry},
        &widget.FormItem{Text: "Password", Widget: getPasswordEntry},
    )

    // Add the forms and buttons to the content container
    content.Add(addForm)
    content.Add(addButton)
    content.Add(getForm)
    content.Add(getButton)
    content.Add(container.NewHBox(
        quitButton,
        container.NewVBox(),
    ))

    // Set the content of the window
    w.SetContent(content)

    // Resize the window
    w.Resize(fyne.NewSize(400, 200))

    // Show and run the application
    w.ShowAndRun()

    // Load passwords at startup
    passwords, err := loadPasswordsFromFile("passwords.json")
    if err != nil {
        // handle error, e.g., file doesn't exist
    }

    // Load passwords at startup
    loadedPasswords, newErr := loadPasswordsFromFile("passwords.json")
    if newErr != nil {
    // handle error
    }

    // If loadedPasswords is not empty, use it
    if len(loadedPasswords.ServicePasswords) > 0 {
        passwords = loadedPasswords
    }

    // Save passwords before exiting
    err = savePasswordsToFile(passwords, "passwords.json")
    if err != nil {
        // handle error
    }
}
