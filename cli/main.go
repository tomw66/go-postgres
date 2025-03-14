package cli

import (
	"database/sql"
	"fmt"
	"go-postgres/database"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type model struct {
    database *sql.DB
	table table.Model
    confirmDelete bool
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			row := m.table.SelectedRow()
			fmt.Println("Editing row:", row)
			fmt.Println("Enter new values (comma separated): ")
			var input string
			fmt.Scanln(&input)
			newValues := strings.Split(input, ",")
			if len(newValues) == len(row) {
				rows := m.table.Rows()
				rows[m.table.Cursor()] = newValues
                m.updateRow(newValues)
				m.table.UpdateViewport()
			} else {
				fmt.Println("Invalid input")
			}
        case "backspace":
            fmt.Println("debug: pressed backspace")
            if m.confirmDelete {
                fmt.Println("Deleting row!")
                m.confirmDelete = false
            } else {
                m.confirmDelete = true
            }
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) updateRow(row []string) {
    var id int
    var name string
    var age int

    if len(row) == 3 {
        id, _ = strconv.Atoi(row[0])
        name = row[1]
        age, _ = strconv.Atoi(row[2])
    }
    database.UpdateRecord(m.database, id, name, age)
    fmt.Println("updated")
}

func (m model) View() string {
    msg := ""
	if m.confirmDelete {
		msg = "Delete row? Press again to confirm\n"
	}
	return fmt.Sprintf("%s%s", msg, m.table.View())
}

func createTable(records []database.Record) table.Model {
    columns := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Name", Width: 10},
		{Title: "Age", Width: 5},
	}

    var rows []table.Row
    for _, record := range records {
        rows = append(rows, table.Row{fmt.Sprintf("%d", record.ID), record.Name, fmt.Sprintf("%d", record.Age)})
    }

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
    fmt.Println("CLI call working!")

    // initialise db
    db := database.InitialiseTable()

    // Read records
	records, err := database.ReadRecords(db)
	if err != nil {
		log.Fatal("Error reading records: ", err)
	}

    table := createTable(records)

	m := model{db, table, false}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}