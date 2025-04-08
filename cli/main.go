package cli

import (
	"database/sql"
	"fmt"
	"go-postgres/database"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	database      *sql.DB
	table         table.Model
	textInput     textinput.Model
	addingRow     bool
	editingRow    bool
	editIndex     int
	message       string
	confirmDelete bool
}

func (m model) Init() tea.Cmd { return nil }

func (m *model) handleTableInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "enter":
		m.confirmDelete = false
		// Add row
		if m.table.Cursor() == len(m.table.Rows())-1 {
			m.addingRow = true
			m.textInput.SetValue("")
			m.textInput.Focus()
			m.message = "✏️ Adding row. Enter new values:"
		} else {
			// Edit row
			m.editIndex = m.table.Cursor()
			m.editingRow = true
			m.textInput.SetValue(strings.Join(m.table.SelectedRow(), ","))
			m.textInput.Focus()
			m.message = fmt.Sprintf("✏️ Editing row %d. Enter new values:", m.editIndex)
		}
		return m, textinput.Blink

	case "backspace":
		if m.confirmDelete {
			rows := m.table.Rows()
			cursor := m.table.Cursor()
			if len(rows) > 0 {
				rows = append(rows[:cursor], rows[cursor+1:]...)
				m.table.SetRows(rows)

				if cursor >= len(rows) && len(rows) > 0 {
					m.table.SetCursor(len(rows) - 1)
				}

				m.message = "✅ Row deleted successfully."
				//TODO deleting wrong record! What's the pattern?
				id, _ := strconv.Atoi(rows[cursor][0])
				err := database.DeleteRecord(m.database,id)
				if err != nil {
					m.message = fmt.Sprintf("❌ Error deleting record: %v", err)
				}
			}
			m.confirmDelete = false
		} else {
			m.message = "Press backspace again to confirm deletion."
			m.confirmDelete = true
		}
		return m, nil

	default:
		m.confirmDelete = false
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *model) addRow(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.Type {
	case tea.KeyEnter:
		input := m.textInput.Value()
		newValues := strings.Split(strings.TrimSpace(input), ",")

		if len(newValues) == 3 {
			rows := m.table.Rows()
			rows = slices.Insert(rows, len(rows)-1 ,newValues)
			m.table.SetRows(rows)
			m.table.SetCursor(len(rows)-2)

			m.message = "✅ Row added successfully."
			age, _ := strconv.Atoi(newValues[2])
			_, err := database.CreateRecord(m.database, newValues[1], age)
			if err != nil {
				m.message = fmt.Sprintf("❌ Error creating record: %v", err)
			}
		} else {
			m.message = "❌ Invalid input. Format: id,name,age"
		}
		m.addingRow = false
		m.textInput.Reset()
		return m, nil

	case tea.KeyEsc:
		// Cancel adding the row
		m.message = "❌ Row addition cancelled."
		m.addingRow = false
		m.textInput.Reset()
		return m, nil
	}
	// Update text input as the user types
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m *model) editRow(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.Type {
	case tea.KeyEnter:
		input := m.textInput.Value()
		newValues := strings.Split(strings.TrimSpace(input), ",")
		row := m.table.SelectedRow()

		if len(newValues) == len(row) {
			rows := m.table.Rows()
			rows[m.editIndex] = newValues
			m.table.SetRows(rows)
			m.table.SetCursor(m.editIndex)

			m.message = "✅ Row updated successfully."
			id, _ := strconv.Atoi(newValues[0])
			age, _ := strconv.Atoi(newValues[2])
			err := database.UpdateRecord(m.database, id, newValues[1], age)
			if err != nil {
				m.message = fmt.Sprintf("❌ Error deleting record: %v", err)
			}
		} else {
			m.message = "❌ Invalid input. Format: id,name,age"
		}
		m.editingRow = false
		m.textInput.Reset()
		return m, nil

	case tea.KeyEsc:
		m.message = "❌ Edit canceled."
		m.editingRow = false
		m.textInput.Reset()
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle adding new row
		if m.addingRow {
			return m.addRow(msg)
		}
		// Handle editing row
		if m.editingRow {
			return m.editRow(msg)
		}
		// Handle normal table input
		return m.handleTableInput(msg)
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString(m.table.View())
	b.WriteString("\n\n")

	if m.editingRow || m.addingRow {
		b.WriteString(m.message + "\n")
		b.WriteString(m.textInput.View())
	} else if m.message != "" {
		b.WriteString(m.message + "\n")
	}

	return b.String()
}

func createTable(records []database.Record) table.Model {
	columns := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Name", Width: 13},
		{Title: "Age", Width: 5},
	}

	var rows []table.Row
	for _, record := range records {
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", record.ID),
			record.Name,
			fmt.Sprintf("%d", record.Age),
		})
	}

	// Add a special row at the end for "Add New Row"
	rows = append(rows, table.Row{"...", "Add New Row", ""})

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return t
}

func CLI() {
    // initialise db
    db := database.InitialiseTable()

    // Read records
	records, err := database.ReadRecords(db)
	if err != nil {
		log.Fatal("Error reading records: ", err)
	}

	ti := textinput.New()
	ti.Placeholder = "id,name,age"
	ti.CharLimit = 100
	ti.Width = 30

	m := model{
		database: db,
		table:     createTable(records),
		textInput: ti,
	}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}