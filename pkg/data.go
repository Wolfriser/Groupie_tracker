package pkg

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	artistAPI    = "https://groupietrackers.herokuapp.com/api/artists"
	relationAPI  = "https://groupietrackers.herokuapp.com/api/relation"
	locationsAPI = "https://groupietrackers.herokuapp.com/api/locations"
)

var (
	ResponseData                               Data
	BandInfo                                   []Band
	RelationInfo                               Relations
	LocationInfo                               Location
	bandInfoMu, relationInfoMu, locationInfoMu sync.RWMutex
)

func SaveCacheToFile(filename string, data []byte) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func UpdateCache() error {
	var err error

	newBandInfo, err := GetBandInfo(artistAPI)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	bandInfoMu.Lock()

	BandInfo = newBandInfo

	defer bandInfoMu.Unlock()

	newRelationInfo, err := GetRelationsInfo(relationAPI)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	relationInfoMu.Lock()

	RelationInfo = newRelationInfo

	defer relationInfoMu.Unlock()

	newLocationInfo, err := GetLocationsInfo(locationsAPI)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	locationInfoMu.Lock()

	LocationInfo = newLocationInfo

	defer locationInfoMu.Unlock()

	AddLocationsToBand(BandInfo, LocationInfo, RelationInfo)
	ResponseData = FillData(BandInfo)

	log.Println("Кэш обновлен")

	return nil
}

// Функция поиска данных в системе данных
func SearchRecords(records []Band, query string) (*[]Band, error) {
	sliceBand := make([]Band, 0)

	if query == "" {
		return nil, fmt.Errorf("Пустой запрос")
	}

	query = removeWords(query)

	for _, record := range records {
		if strings.Contains(strings.ToLower(record.Name), strings.ToLower(query)) ||
			SearchMembers(record.Members, query) ||
			strings.Contains(strings.ToLower(record.FirstAlbum), strings.ToLower(query)) && len(strings.ToLower(record.FirstAlbum)) == len(strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(ConvertToString(record.CreationDate)), strings.ToLower(query)) {
			sliceBand = append(sliceBand, record)
			// return &sliceBand, nil
		}
	}

	if len(sliceBand) > 0 {
		return &sliceBand, nil
	}

	for _, record := range records {
		if SearchLocations(record.Locations, query) {
			sliceBand = append(sliceBand, record)
		}
	}

	if len(sliceBand) > 0 {
		return &sliceBand, nil
	}

	return nil, fmt.Errorf("Поиск c запросом %v не дал результатов", query)
}

// Функция поиска данных об участнике группы
func SearchMembers(slice []string, query string) bool {
	for _, s := range slice {
		if strings.Contains(strings.ToLower(s), strings.ToLower(query)) {
			return true
		}
	}
	return false
}

// Функция поиска локации
func SearchLocations(slice []string, query string) bool {
	for _, s := range slice {
		if strings.Contains(strings.ToLower(s), strings.ToLower(query)) {
			return true
		}
	}
	return false
}

// Функция конвертации строки в число
func ConvertToString(i int) string {
	return strconv.Itoa(i)
}

// Функция которая удаляет ненужные слова
func removeWords(query string) string {
	w := [7]string{"Name: ", "Image: ", "Member: ", "Creation Date: ", "First Album: ", "Location: ", "Concert Date: "}

	for _, word := range w {
		query = strings.ReplaceAll(query, word, "")
	}
	return query
}

// Функция для объединения с Locations
func AddLocationsToBand(band []Band, loc Location, relations Relations) {
	for i, b := range band {
		b.Relations = relations.Index[i].DatesLocations
		b.Locations = loc.Index[i].Locations
		band[i] = b
	}
	return
}

// Функция для получения уникальных локаций из набора данных о группах
func uniqueLocations(bands []Band) []string {
	var locationSet []string
	var uniqueLocations []string
	for _, band := range bands {
		for _, location := range band.Locations {
			locationSet = append(locationSet, location)
		}
	}
	for _, location := range locationSet {
		if repeatString(uniqueLocations, location) != true {
			uniqueLocations = append(uniqueLocations, location)
		}
	}
	return uniqueLocations
}

func repeatString(uq []string, loc string) bool {
	for _, check := range uq {
		if check == loc {
			return true
		}
	}
	return false
}

func repeatInt(uq []int, date int) bool {
	for _, check := range uq {
		if check == date {
			return true
		}
	}
	return false
}

func allNames(bands []Band) []string {
	var Names []string
	for _, band := range bands {
		Names = append(Names, band.Name)
	}
	return Names
}

func allCreationDates(bands []Band) []int {
	var CreationDates []int
	var uniqueDates []int
	for _, band := range bands {
		CreationDates = append(CreationDates, band.CreationDate)
	}
	for _, date := range CreationDates {
		if repeatInt(uniqueDates, date) != true {
			uniqueDates = append(uniqueDates, date)
		}
	}
	return uniqueDates
}

func allFirstAlbums(bands []Band) []string {
	var FirstAlbums []string
	for _, band := range bands {
		FirstAlbums = append(FirstAlbums, band.FirstAlbum)
	}
	return FirstAlbums
}

func allMembers(bands []Band) []string {
	var Members []string
	for _, band := range bands {
		for _, member := range band.Members {
			Members = append(Members, member)
		}
	}
	return Members
}

func FillData(bandinfo []Band) Data {
	ResponseData.Search.Locations = uniqueLocations(bandinfo)
	ResponseData.Search.Names = allNames(bandinfo)
	ResponseData.Search.CreationDates = allCreationDates(bandinfo)
	ResponseData.Search.FirstAlbums = allFirstAlbums(bandinfo)
	ResponseData.Search.Members = allMembers(bandinfo)
	ResponseData.Band = bandinfo

	return ResponseData
}
