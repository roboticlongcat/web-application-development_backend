package repository

import (
	"fmt"
	"strings"
)

type Repository struct {
}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

type Patient struct {
	ID          int
	Name        string
	Sensitivity float32
	Type        int
	Glucose     float32
}

type InsulinCalculation struct {
	ID      int
	Comment string
	Items   []PatientsInsulinCalculation
}

type PatientsInsulinCalculation struct {
	PatientID         int
	CurrentGlucose    float32
	BreadUnits        float32
	CalculatedInsulin float32
}

func (r *Repository) GetPatients() ([]Patient, error) {
	patients := []Patient{
		{
			ID:          1,
			Name:        "Нефедова Екатерина",
			Sensitivity: 1.5,
			Type:        2,
			Glucose:     7,
		},
		{
			ID:          2,
			Name:        "Пушкина Светлана",
			Sensitivity: 2.75,
			Type:        1,
			Glucose:     10,
		},
		{
			ID:          3,
			Name:        "Четкин Вячеслав",
			Sensitivity: 0.5,
			Type:        3,
			Glucose:     3.9,
		},
		{
			ID:          4,
			Name:        "Быстров Дмитрий",
			Sensitivity: 1.2,
			Type:        2,
			Glucose:     5,
		},
		{
			ID:          5,
			Name:        "Забелина Майя",
			Sensitivity: 3.0,
			Type:        1,
			Glucose:     12,
		},
	}

	if len(patients) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}

	return patients, nil
}

func (r *Repository) GetPatient(id int) (Patient, error) {

	patients, err := r.GetPatients()
	if err != nil {
		return Patient{}, err
	}

	for _, patient := range patients {
		if patient.ID == id {
			return patient, nil
		}
	}
	return Patient{}, fmt.Errorf("пациент не найден")
}

func (r *Repository) GetPatientsByName(name string) ([]Patient, error) {
	patients, err := r.GetPatients()
	if err != nil {
		return []Patient{}, err
	}

	var result []Patient
	for _, patient := range patients {
		if strings.Contains(strings.ToLower(patient.Name), strings.ToLower(name)) {
			result = append(result, patient)
		}
	}

	return result, nil
}

func (r *Repository) GetInsulinCalculation(insulinCalculationID int) (InsulinCalculation, error) {
	insulinCalculation := InsulinCalculation{
		ID:      insulinCalculationID,
		Comment: "как можно скорее!!",
		Items: []PatientsInsulinCalculation{
			{PatientID: 5, CurrentGlucose: 12, BreadUnits: 5, CalculatedInsulin: 6},
			{PatientID: 1, CurrentGlucose: 10, BreadUnits: 10, CalculatedInsulin: 5},
		},
	}
	return insulinCalculation, nil
}

func (r *Repository) GetInsulinCalculationItemsCount(insulinCalculationID int) (int, error) {
	calculation, err := r.GetInsulinCalculation(insulinCalculationID)
	if err != nil {
		return 0, err
	}

	return len(calculation.Items), nil
}
