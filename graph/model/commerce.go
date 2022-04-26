package model

// Ici on utilise un modèle différent de celui généré car on
// ne veut pas générer une requette pour certaines choses comme
// le commerçant, les produits ou les commentaires temps
// qu'ils ne sont pas demandés.

func (input *ScheduleInput) ToModel() *Schedule {
	if input == nil {
		return nil
	}

	return &Schedule{
		Opening: input.Opening,
		Closing: input.Closing,
	}
}

func (input *NewBusinessHours) ToModel() *BusinessHours {
	if input == nil {
		return nil
	}

	var mondaySchedules []*Schedule
	var tuesdaySchedules []*Schedule
	var wednesdaySchedules []*Schedule
	var thursdaySchedules []*Schedule
	var fridaySchedules []*Schedule
	var saturdaySchedules []*Schedule
	var sundaySchedules []*Schedule

	for _, monday := range input.Monday {
		mondaySchedules = append(mondaySchedules, monday.ToModel())
	}

	for _, tuesday := range input.Tuesday {
		tuesdaySchedules = append(tuesdaySchedules, tuesday.ToModel())
	}

	for _, wednesday := range input.Wednesday {
		wednesdaySchedules = append(wednesdaySchedules, wednesday.ToModel())
	}

	for _, thursay := range input.Thursday {
		thursdaySchedules = append(thursdaySchedules, thursay.ToModel())
	}

	for _, friday := range input.Friday {
		fridaySchedules = append(fridaySchedules, friday.ToModel())
	}

	for _, saturday := range input.Saturday {
		saturdaySchedules = append(saturdaySchedules, saturday.ToModel())
	}

	for _, sunday := range input.Sunday {
		sundaySchedules = append(sundaySchedules, sunday.ToModel())
	}
	return &BusinessHours{
		Monday:    mondaySchedules,
		Tuesday:   tuesdaySchedules,
		Wednesday: wednesdaySchedules,
		Thursday:  thursdaySchedules,
		Friday:    fridaySchedules,
		Saturday:  saturdaySchedules,
		Sunday:    sundaySchedules,
	}
}

type Commerce struct {
	ID              string        `json:"id"`
	StorekeeperID   string        `json:"storekeeper"`
	Name            string        `json:"name"`
	Description     string        `json:"description"`
	StorekeeperWord string        `json:"storekeeperWord"`
	Address         string        `json:"address"`
	Latitude        float64       `json:"latitude"`
	Longitude       float64       `json:"longitude"`
	Phone           string        `json:"phone"`
	Email           string        `json:"email"`
	Facebook        *string       `json:"facebook"`
	Twitter         *string       `json:"twitter"`
	Instagram       *string       `json:"instagram"`
	BusinessHours   BusinessHours `json:"businessHours"`
	Services        []string      `json:"services"`
}
