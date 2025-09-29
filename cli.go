// This is all AI generated slop because I was too lazy to make it myself

package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/DannyZolp/dns/management"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#000")).
			Background(lipgloss.Color("#FFF")).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#000")).
			Background(lipgloss.Color("#FFF")).
			Padding(0, 1)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF")).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#000")).
			Background(lipgloss.Color("#FFF")).
			Padding(0, 2)

	recordStyle = lipgloss.NewStyle().
			Padding(0, 1)

	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF0000")).
			Padding(0, 1)
)

type state int

const (
	zoneSelection state = iota
	recordTypeSelection
	viewRecords
	addRecord
	editRecord
	deleteRecord
	confirmDelete
	recordInput
	addZone
	editZone
	deleteZone
	confirmDeleteZone
)

type recordType int

const (
	aRecord recordType = iota
	aaaaRecord
	cnameRecord
	mxRecord
	txtRecord
	soaRecord
)

type model struct {
	state          state
	cursor         int
	db             *gorm.DB
	recordType     recordType
	records        interface{}
	selectedRecord interface{}
	inputFields    map[string]string
	currentField   string
	fieldIndex     int
	fieldNames     []string
	error          string
	confirmAction  string
	zones          []management.Zone
	selectedZone   *management.Zone
	zoneID         int
}

type recordDisplay struct {
	ID    string
	Type  string
	Name  string
	Value string
	TTL   string
}

func initialModel() model {
	db, err := gorm.Open(sqlite.Open("dns.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// Auto-migrate the schema with the new Zone-based models
	db.AutoMigrate(&management.Zone{}, &management.SOA{}, &management.A{}, &management.AAAA{}, &management.CNAME{}, &management.MX{}, &management.TXT{})

	m := model{
		state:       zoneSelection,
		db:          db,
		inputFields: make(map[string]string),
	}

	// Load zones initially
	db.Preload("SOA").Find(&m.zones)

	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "esc":
			switch m.state {
			case zoneSelection:
				return m, tea.Quit
			case recordTypeSelection:
				m.loadZones()
				m.state = zoneSelection
				m.cursor = 0
			case viewRecords:
				m.state = recordTypeSelection
				m.cursor = int(m.recordType)
			case recordInput:
				if m.recordType == soaRecord {
					// For SOA records, go back to record type selection since there's no viewRecords
					m.state = recordTypeSelection
					m.cursor = int(m.recordType)
				} else {
					m.loadRecords()
					m.state = viewRecords
					m.cursor = 0
				}
			case confirmDelete, confirmDeleteZone:
				if m.state == confirmDeleteZone {
					m.state = zoneSelection
				} else {
					m.state = viewRecords
				}
				m.cursor = 0
			case addZone, editZone:
				m.loadZones()
				m.state = zoneSelection
				m.cursor = 0
			}
			// Clear form data and errors when navigating back
			m.error = ""
			m.inputFields = make(map[string]string)
			return m, nil

		case "up":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down":
			switch m.state {
			case zoneSelection:
				if m.cursor < len(m.zones) { // Don't go past the "Add Zone" option
					m.cursor++
				}
			case recordTypeSelection:
				if m.cursor < 5 {
					m.cursor++
				}
			case viewRecords:
				records := m.getDisplayRecords()
				if m.cursor < len(records)-1 {
					m.cursor++
				}
			case recordInput, addZone, editZone:
				if m.fieldIndex < len(m.fieldNames)-1 {
					m.fieldIndex++
					m.currentField = m.fieldNames[m.fieldIndex]
				}
			case confirmDelete, confirmDeleteZone:
				if m.cursor < 1 { // Yes/No options
					m.cursor++
				}
			}

		case "enter":
			return m.handleEnter()

		case "d":
			if m.state == recordInput || m.state == addZone || m.state == editZone {
				return m.handleInput(msg.String())
			} else if m.state == viewRecords && len(m.getDisplayRecords()) > 0 && m.recordType != soaRecord {
				// Don't allow deleting SOA records
				m.selectedRecord = m.getDisplayRecords()[m.cursor]
				m.state = confirmDelete
				m.confirmAction = "delete"
				return m, nil
			} else if m.state == zoneSelection && len(m.zones) > 0 && m.cursor < len(m.zones) {
				m.selectedZone = &m.zones[m.cursor]
				m.state = confirmDeleteZone
				m.cursor = 0
				return m, nil
			}

		case "e":
			if m.state == recordInput || m.state == addZone || m.state == editZone {
				return m.handleInput(msg.String())
			} else if m.state == viewRecords && len(m.getDisplayRecords()) > 0 {
				m.selectedRecord = m.getDisplayRecords()[m.cursor]
				m.setupEditForm()
				m.state = recordInput
				return m, nil
			} else if m.state == zoneSelection && len(m.zones) > 0 && m.cursor < len(m.zones) {
				m.selectedZone = &m.zones[m.cursor]
				m.setupEditZoneForm()
				m.state = editZone
				return m, nil
			}

		case "a":
			if m.state == recordInput || m.state == addZone || m.state == editZone {
				return m.handleInput(msg.String())
			} else if m.state == viewRecords && m.recordType != soaRecord {
				// Don't allow adding new SOA records
				m.setupAddForm()
				m.state = recordInput
				return m, nil
			} else if m.state == viewRecords && m.recordType == soaRecord {
				// For SOA records, redirect to edit the existing one
				m.setupAddForm() // This will automatically redirect to edit mode
				m.state = recordInput
				return m, nil
			} else if m.state == zoneSelection {
				m.setupAddZoneForm()
				m.state = addZone
				return m, nil
			}

		default:
			if m.state == recordInput || m.state == addZone || m.state == editZone {
				return m.handleInput(msg.String())
			}
		}
	}

	return m, nil
}

func (m model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.state {
	case zoneSelection:
		if m.cursor < len(m.zones) {
			// Select a zone
			m.selectedZone = &m.zones[m.cursor]
			m.zoneID = int(m.zones[m.cursor].ID)
			m.state = recordTypeSelection
			m.cursor = 0
		} else {
			// Add new zone option
			m.setupAddZoneForm()
			m.state = addZone
		}

	case recordTypeSelection:
		m.recordType = recordType(m.cursor)
		if m.recordType == soaRecord {
			// For SOA records, bypass the list view and go directly to edit
			m.setupSOAEditForm()
			m.state = recordInput
		} else {
			m.loadRecords()
			m.state = viewRecords
			m.cursor = 0
		}

	case viewRecords:
		// Enter on a record could show details or edit
		if len(m.getDisplayRecords()) > 0 {
			if m.recordType == soaRecord {
				// For SOA records, load directly from database instead of using display record
				m.setupSOAEditForm()
			} else {
				m.selectedRecord = m.getDisplayRecords()[m.cursor]
				m.setupEditForm()
			}
			m.state = recordInput
		}

	case recordInput:
		return m.handleFormSubmission()

	case addZone, editZone:
		return m.handleZoneFormSubmission()

	case confirmDelete:
		if m.cursor == 0 { // Yes
			m.deleteSelectedRecord()
			m.loadRecords()
			m.state = viewRecords
			m.cursor = 0
		} else { // No
			m.state = viewRecords
		}

	case confirmDeleteZone:
		if m.cursor == 0 { // Yes
			m.deleteSelectedZone()
			m.loadZones()
			m.state = zoneSelection
			m.cursor = 0
		} else { // No
			m.state = zoneSelection
		}
	}

	return m, nil
}

func (m *model) handleInput(key string) (tea.Model, tea.Cmd) {
	if m.currentField == "" && len(m.fieldNames) > 0 {
		m.currentField = m.fieldNames[0]
	}

	switch key {
	case "backspace":
		if len(m.inputFields[m.currentField]) > 0 {
			m.inputFields[m.currentField] = m.inputFields[m.currentField][:len(m.inputFields[m.currentField])-1]
		}
	case "tab":
		if m.fieldIndex < len(m.fieldNames)-1 {
			m.fieldIndex++
		} else {
			m.fieldIndex = 0
		}
		m.currentField = m.fieldNames[m.fieldIndex]
	default:
		if len(key) == 1 {
			m.inputFields[m.currentField] += key
		}
	}

	return *m, nil
}

func (m *model) handleFormSubmission() (tea.Model, tea.Cmd) {
	// Validate required fields
	for _, field := range m.fieldNames {
		if field != "ID" && m.inputFields[field] == "" {
			m.error = fmt.Sprintf("Field '%s' is required", field)
			return *m, nil
		}
	}

	var err error
	if m.inputFields["ID"] != "" && m.inputFields["ID"] != "0" {
		// Update existing record
		err = m.updateRecord()
	} else {
		// Create new record
		err = m.createRecord()
	}

	if err != nil {
		m.error = err.Error()
		return *m, nil
	}

	// Success - return to appropriate state
	if m.recordType == soaRecord {
		// For SOA records, return to record type selection
		m.state = recordTypeSelection
		m.cursor = int(m.recordType)
	} else {
		// For other records, return to view records
		m.loadRecords()
		m.state = viewRecords
		m.cursor = 0
	}
	m.inputFields = make(map[string]string)
	m.error = ""

	return *m, nil
}

func (m *model) setupAddForm() {
	m.inputFields = make(map[string]string)
	m.fieldIndex = 0
	m.error = ""

	switch m.recordType {
	case aRecord:
		m.fieldNames = []string{"Name", "IP", "TTL"}
		m.currentField = "Name"
	case aaaaRecord:
		m.fieldNames = []string{"Name", "IP", "TTL"}
		m.currentField = "Name"
	case cnameRecord:
		m.fieldNames = []string{"Name", "Target", "TTL"}
		m.currentField = "Name"
	case mxRecord:
		m.fieldNames = []string{"Name", "Target", "Priority", "TTL"}
		m.currentField = "Name"
	case txtRecord:
		m.fieldNames = []string{"Name", "Content", "TTL"}
		m.currentField = "Name"
	case soaRecord:
		// For SOA records, we don't allow adding new ones - only editing existing
		// Use the dedicated SOA edit form setup
		m.setupSOAEditForm()
		return
	}

	// Set default values
	for _, field := range m.fieldNames {
		if field == "TTL" {
			m.inputFields[field] = "300"
		} else if field == "Priority" {
			m.inputFields[field] = "10"
		} else if field == "SerialNumber" {
			m.inputFields[field] = "1"
		} else if field == "Refresh" {
			m.inputFields[field] = "3600"
		} else if field == "Retry" {
			m.inputFields[field] = "1800"
		} else if field == "Expire" {
			m.inputFields[field] = "604800"
		} else {
			m.inputFields[field] = ""
		}
	}
}

func (m *model) setupEditForm() {
	m.inputFields = make(map[string]string)
	m.fieldIndex = 0
	m.error = ""

	record := m.selectedRecord.(recordDisplay)

	switch m.recordType {
	case aRecord:
		m.fieldNames = []string{"ID", "Name", "IP", "TTL"}
	case aaaaRecord:
		m.fieldNames = []string{"ID", "Name", "IP", "TTL"}
	case cnameRecord:
		m.fieldNames = []string{"ID", "Name", "Target", "TTL"}
	case mxRecord:
		m.fieldNames = []string{"ID", "Name", "Target", "Priority", "TTL"}
	case txtRecord:
		m.fieldNames = []string{"ID", "Name", "Content", "TTL"}
	case soaRecord:
		m.fieldNames = []string{"ID", "SecondLevelDomain", "SerialNumber", "TTL", "Refresh", "Retry", "Expire"}
	}

	m.currentField = m.fieldNames[1] // Skip ID field for editing

	// Pre-populate with existing values
	m.inputFields["ID"] = record.ID
	switch m.recordType {
	case aRecord, aaaaRecord:
		m.inputFields["Name"] = record.Name
		m.inputFields["IP"] = record.Value
		m.inputFields["TTL"] = record.TTL
	case cnameRecord:
		m.inputFields["Name"] = record.Name
		m.inputFields["Target"] = record.Value
		m.inputFields["TTL"] = record.TTL
	case mxRecord:
		m.inputFields["Name"] = record.Name
		parts := strings.Fields(record.Value)
		if len(parts) >= 2 {
			m.inputFields["Priority"] = parts[0]
			m.inputFields["Target"] = strings.Join(parts[1:], " ")
		}
		m.inputFields["TTL"] = record.TTL
	case txtRecord:
		m.inputFields["Name"] = record.Name
		m.inputFields["Content"] = record.Value
		m.inputFields["TTL"] = record.TTL
	case soaRecord:
		// Load the actual SOA record to get all fields
		id, _ := strconv.ParseUint(record.ID, 10, 32)
		var soa management.SOA
		if err := m.db.First(&soa, id).Error; err == nil {
			m.inputFields["SecondLevelDomain"] = soa.SecondLevelDomain
			m.inputFields["SerialNumber"] = fmt.Sprintf("%d", soa.SerialNumber)
			m.inputFields["TTL"] = fmt.Sprintf("%d", soa.TTL)
			m.inputFields["Refresh"] = fmt.Sprintf("%d", soa.Refresh)
			m.inputFields["Retry"] = fmt.Sprintf("%d", soa.Retry)
			m.inputFields["Expire"] = fmt.Sprintf("%d", soa.Expire)
		}
	}
}

func (m *model) setupSOAEditForm() {
	m.inputFields = make(map[string]string)
	m.fieldIndex = 0
	m.error = ""

	// Load the SOA record directly from database for this zone
	var soa management.SOA
	if err := m.db.Where("zone_id = ?", m.zoneID).First(&soa).Error; err != nil {
		m.error = "No SOA record found for this zone"
		return
	}

	// Set up the form fields for SOA editing
	m.fieldNames = []string{"ID", "SecondLevelDomain", "SerialNumber", "TTL", "Refresh", "Retry", "Expire"}
	m.currentField = m.fieldNames[1] // Skip ID field for editing

	// Create a selectedRecord for consistency with other record types
	m.selectedRecord = recordDisplay{
		ID:   fmt.Sprintf("%d", soa.ID),
		Type: "SOA",
		Name: soa.SecondLevelDomain,
		Value: fmt.Sprintf("Serial: %d, Refresh: %d, Retry: %d, Expire: %d",
			soa.SerialNumber, soa.Refresh, soa.Retry, soa.Expire),
		TTL: fmt.Sprintf("%d", soa.TTL),
	}

	// Pre-populate with existing SOA values
	m.inputFields["ID"] = fmt.Sprintf("%d", soa.ID)
	m.inputFields["SecondLevelDomain"] = soa.SecondLevelDomain
	m.inputFields["SerialNumber"] = fmt.Sprintf("%d", soa.SerialNumber)
	m.inputFields["TTL"] = fmt.Sprintf("%d", soa.TTL)
	m.inputFields["Refresh"] = fmt.Sprintf("%d", soa.Refresh)
	m.inputFields["Retry"] = fmt.Sprintf("%d", soa.Retry)
	m.inputFields["Expire"] = fmt.Sprintf("%d", soa.Expire)
}

func (m *model) createRecord() error {
	var err error

	switch m.recordType {
	case aRecord:
		ttl, _ := strconv.ParseUint(m.inputFields["TTL"], 10, 32)
		record := management.A{
			Name:   m.inputFields["Name"],
			IP:     m.inputFields["IP"],
			TTL:    uint32(ttl),
			ZoneID: m.zoneID,
		}
		err = m.db.Create(&record).Error

	case aaaaRecord:
		ttl, _ := strconv.ParseUint(m.inputFields["TTL"], 10, 32)
		record := management.AAAA{
			Name:   m.inputFields["Name"],
			IP:     m.inputFields["IP"],
			TTL:    uint32(ttl),
			ZoneID: m.zoneID,
		}
		err = m.db.Create(&record).Error

	case cnameRecord:
		ttl, _ := strconv.ParseUint(m.inputFields["TTL"], 10, 32)
		record := management.CNAME{
			Name:   m.inputFields["Name"],
			Target: m.inputFields["Target"],
			TTL:    uint32(ttl),
			ZoneID: m.zoneID,
		}
		err = m.db.Create(&record).Error

	case mxRecord:
		ttl, _ := strconv.ParseUint(m.inputFields["TTL"], 10, 32)
		priority, _ := strconv.ParseUint(m.inputFields["Priority"], 10, 16)
		record := management.MX{
			Name:     m.inputFields["Name"],
			Target:   m.inputFields["Target"],
			Priority: uint16(priority),
			TTL:      uint32(ttl),
			ZoneID:   m.zoneID,
		}
		err = m.db.Create(&record).Error

	case txtRecord:
		ttl, _ := strconv.ParseUint(m.inputFields["TTL"], 10, 32)
		record := management.TXT{
			Name:    m.inputFields["Name"],
			Content: m.inputFields["Content"],
			TTL:     uint32(ttl),
			ZoneID:  m.zoneID,
		}
		err = m.db.Create(&record).Error

	case soaRecord:
		// Check if SOA already exists for this zone
		var count int64
		m.db.Model(&management.SOA{}).Where("zone_id = ?", m.zoneID).Count(&count)
		if count > 0 {
			return fmt.Errorf("SOA record already exists for this zone. Use edit to modify it")
		}

		ttl, _ := strconv.ParseUint(m.inputFields["TTL"], 10, 32)
		serialNumber, _ := strconv.ParseUint(m.inputFields["SerialNumber"], 10, 32)
		refresh, _ := strconv.ParseUint(m.inputFields["Refresh"], 10, 32)
		retry, _ := strconv.ParseUint(m.inputFields["Retry"], 10, 32)
		expire, _ := strconv.ParseUint(m.inputFields["Expire"], 10, 32)
		record := management.SOA{
			SecondLevelDomain: m.inputFields["SecondLevelDomain"],
			SerialNumber:      uint32(serialNumber),
			TTL:               uint32(ttl),
			Refresh:           uint32(refresh),
			Retry:             uint32(retry),
			Expire:            uint32(expire),
			ZoneID:            m.zoneID,
		}
		err = m.db.Create(&record).Error
	}

	// If the record was created successfully and it's not a SOA record, increment the SOA serial
	if err == nil && m.recordType != soaRecord {
		m.incrementSOASerial()
	}

	return err
}

func (m *model) updateRecord() error {
	id, _ := strconv.ParseUint(m.inputFields["ID"], 10, 32)
	var err error

	switch m.recordType {
	case aRecord:
		ttl, _ := strconv.ParseUint(m.inputFields["TTL"], 10, 32)
		updates := management.A{
			Name: m.inputFields["Name"],
			IP:   m.inputFields["IP"],
			TTL:  uint32(ttl),
		}
		err = m.db.Model(&management.A{}).Where("id = ?", id).Updates(updates).Error

	case aaaaRecord:
		ttl, _ := strconv.ParseUint(m.inputFields["TTL"], 10, 32)
		updates := management.AAAA{
			Name: m.inputFields["Name"],
			IP:   m.inputFields["IP"],
			TTL:  uint32(ttl),
		}
		err = m.db.Model(&management.AAAA{}).Where("id = ?", id).Updates(updates).Error

	case cnameRecord:
		ttl, _ := strconv.ParseUint(m.inputFields["TTL"], 10, 32)
		updates := management.CNAME{
			Name:   m.inputFields["Name"],
			Target: m.inputFields["Target"],
			TTL:    uint32(ttl),
		}
		err = m.db.Model(&management.CNAME{}).Where("id = ?", id).Updates(updates).Error

	case mxRecord:
		ttl, _ := strconv.ParseUint(m.inputFields["TTL"], 10, 32)
		priority, _ := strconv.ParseUint(m.inputFields["Priority"], 10, 16)
		updates := management.MX{
			Name:     m.inputFields["Name"],
			Target:   m.inputFields["Target"],
			Priority: uint16(priority),
			TTL:      uint32(ttl),
		}
		err = m.db.Model(&management.MX{}).Where("id = ?", id).Updates(updates).Error

	case txtRecord:
		ttl, _ := strconv.ParseUint(m.inputFields["TTL"], 10, 32)
		updates := management.TXT{
			Name:    m.inputFields["Name"],
			Content: m.inputFields["Content"],
			TTL:     uint32(ttl),
		}
		err = m.db.Model(&management.TXT{}).Where("id = ?", id).Updates(updates).Error

	case soaRecord:
		ttl, _ := strconv.ParseUint(m.inputFields["TTL"], 10, 32)
		serialNumber, _ := strconv.ParseUint(m.inputFields["SerialNumber"], 10, 32)
		refresh, _ := strconv.ParseUint(m.inputFields["Refresh"], 10, 32)
		retry, _ := strconv.ParseUint(m.inputFields["Retry"], 10, 32)
		expire, _ := strconv.ParseUint(m.inputFields["Expire"], 10, 32)
		updates := management.SOA{
			SecondLevelDomain: m.inputFields["SecondLevelDomain"],
			SerialNumber:      uint32(serialNumber),
			TTL:               uint32(ttl),
			Refresh:           uint32(refresh),
			Retry:             uint32(retry),
			Expire:            uint32(expire),
		}
		err = m.db.Model(&management.SOA{}).Where("id = ?", id).Updates(updates).Error
	}

	// If the record was updated successfully and it's not a SOA record, increment the SOA serial
	if err == nil && m.recordType != soaRecord {
		m.incrementSOASerial()
	}

	return err
}

func (m *model) deleteSelectedRecord() {
	record := m.selectedRecord.(recordDisplay)
	id, _ := strconv.ParseUint(record.ID, 10, 32)

	switch record.Type {
	case "A":
		m.db.Delete(&management.A{}, id)
	case "AAAA":
		m.db.Delete(&management.AAAA{}, id)
	case "CNAME":
		m.db.Delete(&management.CNAME{}, id)
	case "MX":
		m.db.Delete(&management.MX{}, id)
	case "TXT":
		m.db.Delete(&management.TXT{}, id)
	case "SOA":
		m.db.Delete(&management.SOA{}, id)
	}

	// If a non-SOA record was deleted, increment the SOA serial
	if record.Type != "SOA" {
		m.incrementSOASerial()
	}
}

func (m *model) incrementSOASerial() {
	// Find and increment the SOA serial number for the current zone
	var soa management.SOA
	if err := m.db.Where("zone_id = ?", m.zoneID).First(&soa).Error; err == nil {
		// Increment the serial number
		soa.SerialNumber++
		// Update the SOA record in the database
		m.db.Save(&soa)
	}
}

func (m *model) loadRecords() {
	switch m.recordType {
	case aRecord:
		var records []management.A
		m.db.Where("zone_id = ?", m.zoneID).Find(&records)
		m.records = records
	case aaaaRecord:
		var records []management.AAAA
		m.db.Where("zone_id = ?", m.zoneID).Find(&records)
		m.records = records
	case cnameRecord:
		var records []management.CNAME
		m.db.Where("zone_id = ?", m.zoneID).Find(&records)
		m.records = records
	case mxRecord:
		var records []management.MX
		m.db.Where("zone_id = ?", m.zoneID).Find(&records)
		m.records = records
	case txtRecord:
		var records []management.TXT
		m.db.Where("zone_id = ?", m.zoneID).Find(&records)
		m.records = records
	case soaRecord:
		var records []management.SOA
		m.db.Where("zone_id = ?", m.zoneID).Find(&records)
		m.records = records
	}
}

func (m *model) getDisplayRecords() []recordDisplay {
	var displays []recordDisplay

	switch m.recordType {
	case aRecord:
		if records, ok := m.records.([]management.A); ok {
			for _, record := range records {
				displays = append(displays, recordDisplay{
					ID:    fmt.Sprintf("%d", record.ID),
					Type:  "A",
					Name:  record.Name,
					Value: record.IP,
					TTL:   fmt.Sprintf("%d", record.TTL),
				})
			}
		}
	case aaaaRecord:
		if records, ok := m.records.([]management.AAAA); ok {
			for _, record := range records {
				displays = append(displays, recordDisplay{
					ID:    fmt.Sprintf("%d", record.ID),
					Type:  "AAAA",
					Name:  record.Name,
					Value: record.IP,
					TTL:   fmt.Sprintf("%d", record.TTL),
				})
			}
		}
	case cnameRecord:
		if records, ok := m.records.([]management.CNAME); ok {
			for _, record := range records {
				displays = append(displays, recordDisplay{
					ID:    fmt.Sprintf("%d", record.ID),
					Type:  "CNAME",
					Name:  record.Name,
					Value: record.Target,
					TTL:   fmt.Sprintf("%d", record.TTL),
				})
			}
		}
	case mxRecord:
		if records, ok := m.records.([]management.MX); ok {
			for _, record := range records {
				displays = append(displays, recordDisplay{
					ID:    fmt.Sprintf("%d", record.ID),
					Type:  "MX",
					Name:  record.Name,
					Value: fmt.Sprintf("%d %s", record.Priority, record.Target),
					TTL:   fmt.Sprintf("%d", record.TTL),
				})
			}
		}
	case txtRecord:
		if records, ok := m.records.([]management.TXT); ok {
			for _, record := range records {
				displays = append(displays, recordDisplay{
					ID:    fmt.Sprintf("%d", record.ID),
					Type:  "TXT",
					Name:  record.Name,
					Value: record.Content,
					TTL:   fmt.Sprintf("%d", record.TTL),
				})
			}
		}
	case soaRecord:
		if records, ok := m.records.([]management.SOA); ok {
			for _, record := range records {
				displays = append(displays, recordDisplay{
					ID:   fmt.Sprintf("%d", record.ID),
					Type: "SOA",
					Name: record.SecondLevelDomain,
					Value: fmt.Sprintf("Serial: %d, Refresh: %d, Retry: %d, Expire: %d",
						record.SerialNumber, record.Refresh, record.Retry, record.Expire),
					TTL: fmt.Sprintf("%d", record.TTL),
				})
			}
		}
	}

	return displays
}

// Zone management functions
func (m *model) loadZones() {
	m.db.Preload("SOA").Find(&m.zones)
}

func (m *model) setupAddZoneForm() {
	m.inputFields = make(map[string]string)
	m.fieldNames = []string{"Name"}
	m.fieldIndex = 0
	m.currentField = "Name"
	m.error = ""
	m.inputFields["Name"] = ""
}

func (m *model) setupEditZoneForm() {
	m.inputFields = make(map[string]string)
	m.fieldNames = []string{"ID", "Name"}
	m.fieldIndex = 1 // Skip ID field for editing
	m.currentField = "Name"
	m.error = ""
	m.inputFields["ID"] = fmt.Sprintf("%d", m.selectedZone.ID)
	m.inputFields["Name"] = m.selectedZone.Name
}

func (m *model) handleZoneFormSubmission() (tea.Model, tea.Cmd) {
	// Validate required fields
	if m.inputFields["Name"] == "" {
		m.error = "Zone name is required"
		return *m, nil
	}

	var err error
	if m.inputFields["ID"] != "" && m.inputFields["ID"] != "0" {
		// Update existing zone
		err = m.updateZone()
	} else {
		// Create new zone
		err = m.createZone()
	}

	if err != nil {
		m.error = err.Error()
		return *m, nil
	}

	// Success - return to zone selection
	m.loadZones()
	m.state = zoneSelection
	m.cursor = 0
	m.inputFields = make(map[string]string)
	m.error = ""

	return *m, nil
}

func (m *model) createZone() error {
	zone := management.Zone{
		Name: m.inputFields["Name"],
	}

	// Create the zone first
	if err := m.db.Create(&zone).Error; err != nil {
		return err
	}

	// Create a default SOA record for the new zone
	soa := management.SOA{
		SecondLevelDomain: m.inputFields["Name"],
		SerialNumber:      1,
		TTL:               300,
		Refresh:           3600,
		Retry:             1800,
		Expire:            604800,
		ZoneID:            int(zone.ID),
	}

	return m.db.Create(&soa).Error
}

func (m *model) updateZone() error {
	id, _ := strconv.ParseUint(m.inputFields["ID"], 10, 32)
	updates := management.Zone{
		Name: m.inputFields["Name"],
	}
	return m.db.Model(&management.Zone{}).Where("id = ?", id).Updates(updates).Error
}

func (m *model) deleteSelectedZone() {
	if m.selectedZone != nil {
		// Delete all records associated with this zone first
		m.db.Where("zone_id = ?", m.selectedZone.ID).Delete(&management.A{})
		m.db.Where("zone_id = ?", m.selectedZone.ID).Delete(&management.AAAA{})
		m.db.Where("zone_id = ?", m.selectedZone.ID).Delete(&management.CNAME{})
		m.db.Where("zone_id = ?", m.selectedZone.ID).Delete(&management.MX{})
		m.db.Where("zone_id = ?", m.selectedZone.ID).Delete(&management.TXT{})
		m.db.Where("zone_id = ?", m.selectedZone.ID).Delete(&management.SOA{})

		// Delete the zone itself
		m.db.Delete(m.selectedZone)
	}
}

func (m model) View() string {
	switch m.state {
	case zoneSelection:
		return m.renderZoneSelection()
	case recordTypeSelection:
		return m.renderRecordTypeSelection()
	case viewRecords:
		return m.renderViewRecords()
	case recordInput:
		return m.renderRecordInput()
	case confirmDelete:
		return m.renderConfirmDelete()
	case addZone, editZone:
		return m.renderZoneInput()
	case confirmDeleteZone:
		return m.renderConfirmDeleteZone()
	default:
		return "Unknown state"
	}
}

func (m model) renderZoneSelection() string {
	s := titleStyle.Render("DNS Zone Management") + "\n\n"

	if len(m.zones) == 0 {
		s += normalStyle.Render("No zones found") + "\n"
	} else {
		for i, zone := range m.zones {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
				s += selectedStyle.Render(fmt.Sprintf("%s %s", cursor, zone.Name)) + "\n"
			} else {
				s += normalStyle.Render(fmt.Sprintf("%s %s", cursor, zone.Name)) + "\n"
			}
		}
	}

	// Add "Create new zone" option
	cursor := " "
	if m.cursor == len(m.zones) {
		cursor = ">"
		s += selectedStyle.Render(fmt.Sprintf("%s + Add New Zone", cursor)) + "\n"
	} else {
		s += normalStyle.Render(fmt.Sprintf("%s + Add New Zone", cursor)) + "\n"
	}

	s += "\n" + normalStyle.Render("Use ↑/↓ to navigate, Enter to select, a: add zone, e: edit zone, d: delete zone, q to quit")
	return s
}

func (m model) renderRecordTypeSelection() string {
	s := titleStyle.Render(fmt.Sprintf("Zone: %s - Select Record Type", m.selectedZone.Name)) + "\n\n"

	types := []string{"A Record", "AAAA Record", "CNAME Record", "MX Record", "TXT Record", "SOA Record"}

	for i, recordType := range types {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
			s += selectedStyle.Render(fmt.Sprintf("%s %s", cursor, recordType)) + "\n"
		} else {
			s += normalStyle.Render(fmt.Sprintf("%s %s", cursor, recordType)) + "\n"
		}
	}

	s += "\n" + normalStyle.Render("Use ↑/↓ to navigate, Enter to select (SOA goes directly to edit), Esc: back to zones")
	return s
}

func (m model) renderViewRecords() string {
	recordTypeName := []string{"A", "AAAA", "CNAME", "MX", "TXT", "SOA"}
	s := titleStyle.Render(fmt.Sprintf("Zone: %s - %s Records", m.selectedZone.Name, recordTypeName[m.recordType])) + "\n\n"

	records := m.getDisplayRecords()
	if len(records) == 0 {
		s += normalStyle.Render("No records found") + "\n"
	} else {
		// Header
		s += headerStyle.Render(fmt.Sprintf("%-5s %-20s %-30s %-10s", "ID", "Name", "Value", "TTL")) + "\n"

		for i, record := range records {
			line := fmt.Sprintf("%-5s %-20s %-30s %-10s", record.ID, record.Name, record.Value, record.TTL)
			if i == m.cursor {
				s += selectedStyle.Render(line) + "\n"
			} else {
				s += recordStyle.Render(line) + "\n"
			}
		}
	}

	// Different instructions for SOA vs other records
	if m.recordType == soaRecord {
		s += "\n" + normalStyle.Render("↑/↓: navigate, Enter/e: edit SOA, Esc: back (SOA cannot be added/deleted)")
	} else {
		s += "\n" + normalStyle.Render("↑/↓: navigate, Enter/e: edit, a: add, d: delete, Esc: back")
	}
	return s
}

func (m model) renderZoneInput() string {
	action := "Add"
	if m.inputFields["ID"] != "" && m.inputFields["ID"] != "0" {
		action = "Edit"
	}

	s := titleStyle.Render(fmt.Sprintf("%s Zone", action)) + "\n\n"

	if m.error != "" {
		s += errorStyle.Render("Error: "+m.error) + "\n\n"
	}

	for i, field := range m.fieldNames {
		if field == "ID" && (m.inputFields["ID"] == "" || m.inputFields["ID"] == "0") {
			continue // Skip ID field when adding new zones
		}

		value := m.inputFields[field]
		if i == m.fieldIndex && field == m.currentField {
			s += selectedStyle.Render(fmt.Sprintf("%s: %s_", field, value)) + "\n"
		} else {
			s += normalStyle.Render(fmt.Sprintf("%s: %s", field, value)) + "\n"
		}
	}

	s += "\n" + normalStyle.Render("Tab: next field, Enter: submit, Esc: cancel")
	return s
}

func (m model) renderConfirmDeleteZone() string {
	s := titleStyle.Render("Confirm Delete Zone") + "\n\n"
	s += normalStyle.Render(fmt.Sprintf("Delete zone '%s' and all its records?", m.selectedZone.Name)) + "\n\n"

	options := []string{"Yes", "No"}
	for i, option := range options {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
			s += selectedStyle.Render(fmt.Sprintf("%s %s", cursor, option)) + "\n"
		} else {
			s += normalStyle.Render(fmt.Sprintf("%s %s", cursor, option)) + "\n"
		}
	}

	s += "\n" + normalStyle.Render("Use ↑/↓ to navigate, Enter to select")
	return s
}

func (m model) renderRecordInput() string {
	action := "Add"
	if m.inputFields["ID"] != "" && m.inputFields["ID"] != "0" {
		action = "Edit"
	}

	recordTypeName := []string{"A", "AAAA", "CNAME", "MX", "TXT", "SOA"}
	s := titleStyle.Render(fmt.Sprintf("%s %s Record", action, recordTypeName[m.recordType])) + "\n\n"

	if m.error != "" {
		s += errorStyle.Render("Error: "+m.error) + "\n\n"
	}

	for i, field := range m.fieldNames {
		if field == "ID" && (m.inputFields["ID"] == "" || m.inputFields["ID"] == "0") {
			continue // Skip ID field when adding new records
		}

		value := m.inputFields[field]
		if i == m.fieldIndex && field == m.currentField {
			s += selectedStyle.Render(fmt.Sprintf("%s: %s_", field, value)) + "\n"
		} else {
			s += normalStyle.Render(fmt.Sprintf("%s: %s", field, value)) + "\n"
		}
	}

	s += "\n" + normalStyle.Render("Tab: next field, Enter: submit, Esc: cancel")
	return s
}

func (m model) renderConfirmDelete() string {
	record := m.selectedRecord.(recordDisplay)
	s := titleStyle.Render("Confirm Delete") + "\n\n"
	s += normalStyle.Render(fmt.Sprintf("Delete %s record '%s'?", record.Type, record.Name)) + "\n\n"

	options := []string{"Yes", "No"}
	for i, option := range options {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
			s += selectedStyle.Render(fmt.Sprintf("%s %s", cursor, option)) + "\n"
		} else {
			s += normalStyle.Render(fmt.Sprintf("%s %s", cursor, option)) + "\n"
		}
	}

	s += "\n" + normalStyle.Render("Use ↑/↓ to navigate, Enter to select")
	return s
}

func cli() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running CLI: %v\n", err)
		os.Exit(1)
	}
}
