package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const debug = false

var Days = [...]string{"Lunes", "Martes", "Miércoles", "Jueves", "Viernes", "Sábado", "Domingo"}
var DaysNoAccent = [...]string{"Lunes", "Martes", "Miercoles", "Jueves", "Viernes", "Sabado", "Domingo"}
var DaysBadParse = [...]string{"Lúnes", "Mártes", "Mi�rcoles", "Jueves", "Viernes", "Sébado", "Domingo"}
var DaysEnglish = [...]string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}

// TYPES
type comision struct {
	label     string
	schedules []schedule
	teachers  []string
}
type schedule struct {
	day   int8 // from 0 to 6
	start timehm
	end   timehm
}

type scheduleCriteria struct {
	maxSuperposition          float32
	maxTotalSuperposition     float32
	maxNumberOfSuperpositions int
	freeDays                  [len(Days)]bool
	minFreeDays               int
}

// Class as in school class
type Class struct {
	num1       int
	num2       int
	name       string
	comisiones []comision
}

type Cursada []comision // Cursada is a the group of courses a student attends during the year/semester

func GetSchedules(classes []*Class, criteria *scheduleCriteria) *[]Cursada {
	currentSchedule := NewCursada()
	scheduleListMaster := searcher(classes, &currentSchedule, 0, criteria)

	// Search function: cada instancia de recursividad busca verificar un schedule (lo va fabricando a medida que avanza) y devuelve una lista de schedules verificados y los va juntando en cada instancia
	if len(*scheduleListMaster) == 0 {
		return nil
	}
	return scheduleListMaster
}

func searcher(classes []*Class, currentCursada *Cursada, classNumber int, criteria *scheduleCriteria) *[]Cursada {
	nextClass := classes[classNumber]
	cursadaListMaster := NewCursadaList()
	for _, v := range nextClass.comisiones {
		//cursadaInstance := new(Cursada)
		cursadaInstance := append(*currentCursada, v)

		if classNumber == len(classes)-1 { //llegue a la ultima clase
			isValid := verifyCursada(&cursadaInstance, criteria)
			if isValid { //El schedule es bueno, lo devuelvo como lista no nula
				cursadaInstanceCopy := make(Cursada, len(cursadaInstance))
				copy(cursadaInstanceCopy, cursadaInstance)
				cursadaListMaster = append(cursadaListMaster, cursadaInstanceCopy)
				continue
			} else {
				continue
			}

		} else { // Si no es la ultima clase, sigo por aca
			cursadaList := searcher(classes, &cursadaInstance, classNumber+1, criteria) // Awesome recursion baby
			if len(*cursadaList) == 0 {
				continue
			}
			cursadaListMaster = append(cursadaListMaster, *cursadaList...)
		}
	}
	return &cursadaListMaster
}

func (mySchedule schedule) HourDuration() float32 {
	return float32(mySchedule.end.hour - mySchedule.start.hour + (mySchedule.end.minute-mySchedule.start.minute)/60)

}

func NewSchedule() schedule {
	return schedule{}
}

type timehm struct {
	hour   int
	minute int
}

func NewTime() timehm {
	return timehm{}
}
func NewComision() comision {
	return comision{}
}

func NewScheduleCriteria() scheduleCriteria {
	return scheduleCriteria{}
}

func NewCursada() Cursada {
	return Cursada{}
}

func NewCursadaList() []Cursada {
	return []Cursada{}
}

func NewClass() Class {
	return Class{}
}

func GatherClasses(filedir string) ([]*Class, error) {
	f, err := os.Open(filedir)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)

	line := 0

	reClassNumber := regexp.MustCompile(`[0-9]{2}\.[0-9]{2}`)
	reSchedule := regexp.MustCompile(`^[\s]{0,99}[A-Za-zéá�]{5,9}[\s][0-9:0-9]{5}[\s-]{1,5}[0-9:0-9]{5}`)
	reComisionLabel := regexp.MustCompile(`(?:^[\s]{0,99})[A-Z]{1,8}(?:[\s]{0,99}$)`)
	reEndComision := regexp.MustCompile(`^[\s]{0,99}[0-9]{1,5}[\/\s]{1,3}[\w]{1,11}[\s]{0,99}$`)
	reEAccent := regexp.MustCompile(`[�]{1}`)

	var (
		//currentClass          Class
		allClasses            []*Class
		numberString          string
		currentStringSchedule string
		comisionAppended      bool
		yetToAppendToClasses  bool
	)
	badChar := '�'
	for scanner.Scan() { // SCANNER SUPERIOR
		line++
		textLine := scanner.Text()

		if len(textLine) == 0 {
			continue
		}
		//Sanitize unicode disgusting badness
		// Si encuentro una clase
		for reClassNumber.MatchString(textLine) {
			if debug {
				fmt.Printf("[DEBUG] Nueva class hallada (%d)\n", line)
			}
			currentClass := NewClass()
			numberString = reClassNumber.FindString(textLine)
			currentClass.num1, err = strconv.Atoi(numberString[0:2])
			if err != nil {
				break
			}
			currentClass.num2, err = strconv.Atoi(numberString[3:5])
			if err != nil {
				break
			}
			currentClass.name = textLine[8:]

			currentComision := NewComision()
			// Entro en el for loop de las comisiones
			for scanner.Scan() {
				line++
				textLine = scanner.Text()
				if reClassNumber.MatchString(textLine) {
					if len(currentClass.comisiones) > 0 {
						allClasses = append(allClasses, &currentClass)
						yetToAppendToClasses = false
					}

					if debug {
						fmt.Printf("[DEBUG] Fin de class y comienzo de otra hallada (%d)\n", line)
					}
					break
				}
				// Si es una comision:
				if reComisionLabel.MatchString(textLine) {
					if debug {
						fmt.Printf("[DEBUG] Nueva Comision %s encontrada (%d)\n", textLine, line)
					}
					if currentComision.label != "" {
						if !comisionAppended {
							currentClass.comisiones = append(currentClass.comisiones, currentComision)
							if debug {
								fmt.Printf("[DEBUG] Comision %s append a class (%d)\n", currentComision.label, line)
							}
						}
						currentComision = NewComision()
					}

					currentComision.label = reComisionLabel.FindString(textLine)

				}

				if reEndComision.MatchString(textLine) {
					currentClass.comisiones = append(currentClass.comisiones, currentComision)
					comisionAppended = true
					yetToAppendToClasses = true
					if debug {
						fmt.Printf("[DEBUG] Fin de una comision. (%d)\n", line)
					}
					continue
				}

				currentStringSchedule = reSchedule.FindString(textLine)

				if currentStringSchedule != "" {
					comisionAppended = false
					// se usaron 3 algoritmos para eliminar los caracteres malos
					currentStringSchedule = reEAccent.ReplaceAllString(currentStringSchedule, "é")
					//currentStringSchedule = strings.Replace(currentStringSchedule,string(badChar),"é",-1)
					for i, s := range currentStringSchedule {
						if s == badChar {
							currentStringSchedule = currentStringSchedule[:i] + "é" + currentStringSchedule[i+1:]
						}
					}

					diaInt, theTime, err := stringToTime(currentStringSchedule)
					if err != nil {
						return nil, err
					}

					if debug {
						fmt.Printf("[DEBUG] DIA: %s STRUCT TIME: %+v\n\n", Days[diaInt], theTime)
					}

					currentSchedule := NewSchedule()
					currentSchedule.start = theTime[0]
					currentSchedule.end = theTime[1]
					currentSchedule.day = diaInt

					currentComision.schedules = append(currentComision.schedules, currentSchedule)
				}
				if strings.Contains(textLine, ",") {
					currentComision.teachers = append(currentComision.teachers, textLine)
				}
			}
			if len(currentClass.comisiones) > 0 && yetToAppendToClasses {
				allClasses = append(allClasses, &currentClass)

			}
		}

	}
	if debug {
		fmt.Printf("[DEBUG] Se termino de buscar Class. GatherClass Over (%d)\n", line)
	}

	if err != nil {
		err = fmt.Errorf("Hubo un error (%d): %s\n", line, err)
	}
	return allClasses, err
}

func findCollision(schedule1 *schedule, schedule2 *schedule) float32 {
	if (schedule1.start.hour >= schedule2.start.hour && schedule1.start.hour < schedule2.end.hour) && schedule1.HourDuration() >= 0.5 {
		if schedule1.end.hour <= schedule2.end.hour {
			return schedule1.HourDuration()
		} else {
			return schedule1.HourDuration() + float32(-schedule1.end.hour+schedule2.end.hour)
		}
	}
	if (schedule2.start.hour >= schedule1.start.hour && schedule2.start.hour < schedule1.end.hour) && schedule2.HourDuration() >= 0.5 {
		if schedule2.end.hour <= schedule1.end.hour {
			return schedule2.HourDuration()
		} else {
			return schedule2.HourDuration() + float32(-schedule2.end.hour+schedule1.end.hour)
		}
	}
	return 0.0
}

func verifyCursada(currentCursada *Cursada, criteria *scheduleCriteria) bool {
	// TODO hard part coming ahead. Actual verification

	numberOfMaterias := len(*currentCursada)
	superpositionCounter := 0
	totalSuperpositions := float32(0)
	var busyDays = []bool{false, false, false, false, false, false, false}
	for i := 0; i < numberOfMaterias-1; i++ {
		firstComision := (*currentCursada)[i]
		for j := numberOfMaterias - 1; j > i; j-- {
			secondComision := (*currentCursada)[j]
			firstCursada := firstComision.schedules
			secondCursada := secondComision.schedules
			for _, schedule1 := range firstCursada {
				for _, schedule2 := range secondCursada {
					busyDays[schedule1.day] = true
					busyDays[schedule2.day] = true
					if criteria.freeDays[schedule1.day] || criteria.freeDays[schedule2.day] { // Criterio absoluto. Si quiero un free day, entonces se va hacer un free day
						return false
					}
					if schedule1.day != schedule2.day { // si no coinciden los dias, verifica ese horario, continuo buscando colisioneschedule1s
						continue
					} else { // en el caso que sean el mismo día:
						superpositions := findCollision(&schedule1, &schedule2)
						if superpositions == 0.0 { // NO se encuentran superposiciones
							continue
						} else {
							superpositionCounter++
							totalSuperpositions += superpositions
							if totalSuperpositions > criteria.maxTotalSuperposition || superpositionCounter > criteria.maxNumberOfSuperpositions || superpositions > criteria.maxSuperposition {
								return false
							}
						}
					}
				}
			}

		}
	}
	var totalBusyDays int
	// Verificación final de dias ocupados
	for i, b := range busyDays {
		if b {
			totalBusyDays++ // Contador de los dias ocupados
			if criteria.freeDays[i] {
				return false
			}
		}
		//if criteria.freeDays[i] && b { // OLD METHOD
		//	return false
		//}
	}
	if totalBusyDays > 5 - criteria.minFreeDays { // Verificacion de minima cantidad de dias libres
		return false
	}

	return true
}

func stringToTime(scheduleString string) (int8, []timehm, error) {
	reWeek := regexp.MustCompile(`(?i)Lunes|Martes|Miércoles|Jueves|Viernes|Sábado|Domingo|Miercoles|Sabado|Sébado|Mi�rcoles`)
	reSchedule := regexp.MustCompile(`[0-9:0-9]{5}[\s-]{1,5}[0-9:0-9]{5}`)
	reScheduleStart := regexp.MustCompile(`^[0-9:0-9]{5}`)
	reScheduleFinish := regexp.MustCompile(`[0-9:0-9]{5}$`)
	reHours := regexp.MustCompile(`^[0-9]{2}`)
	reMinutes := regexp.MustCompile(`[0-9]{2}$`)
	diaString := reWeek.FindString(scheduleString)
	//dayInt := -1
	var dayInt int8
	dayInt = -1
	for i, v := range Days {
		if strings.Contains(strings.Title(diaString), v) {
			dayInt = int8(i)
		} else if strings.Contains(strings.Title(diaString), DaysNoAccent[i]) || strings.Contains(strings.Title(diaString), DaysBadParse[i]) || strings.Contains(strings.Title(diaString), DaysEnglish[i]) {
			dayInt = int8(i)
		}
	}
	if dayInt == -1 {

		return dayInt, nil, fmt.Errorf("Failed to match string  %s  with a day.", diaString)
	}
	timeString := reSchedule.FindString(scheduleString)
	startTime := reScheduleStart.FindString(timeString)
	endTime := reScheduleFinish.FindString(timeString)
	startTimeHours := reHours.FindString(startTime)
	startTimeMinutes := reMinutes.FindString(startTime)
	endTimeHours := reHours.FindString(endTime)
	endTimeMinutes := reMinutes.FindString(endTime)
	timeStart := NewTime()
	timeEnd := NewTime()
	number1, _ := strconv.Atoi(startTimeHours)
	number2, _ := strconv.Atoi(startTimeMinutes)
	timeStart.hour = number1
	timeStart.minute = number2
	number3, _ := strconv.Atoi(endTimeHours)
	number4, _ := strconv.Atoi(endTimeMinutes)

	timeEnd.hour = number3
	timeStart.minute = number4
	return dayInt, []timehm{timeStart, timeEnd}, nil
}

func PossibleCombinations(theClasses []*Class) int {
	var n int
	n = 1
	for _, v := range theClasses {
		n = n * len(v.comisiones)
	}
	return n
}
