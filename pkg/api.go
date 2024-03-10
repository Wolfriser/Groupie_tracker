package pkg

import (
	"encoding/json"
	"net/http"
)

type Data struct {
	Band   []Band
	Search Search
}

type Band struct {
	ID           int                 `json:"id"`
	Image        string              `json:"image"`
	Name         string              `json:"name"`
	Members      []string            `json:"members"`
	CreationDate int                 `json:"creationDate"`
	FirstAlbum   string              `json:"firstAlbum"`
	Locations    []string            `json:"-"`
	ConcertDates string              `json:"concertDates"`
	Relations    map[string][]string `json:"-"`
}

type Relations struct {
	Index []struct {
		ID             int                 `json:"id"`
		DatesLocations map[string][]string `json:"datesLocations"`
	} `json:"index"`
}

type Location struct {
	Index []struct {
		ID        int      `json:"id"`
		Locations []string `json:"locations"`
		Dates     string   `json:"dates"`
	} `json:"index"`
}

type Search struct {
	Names         []string
	CreationDates []int
	FirstAlbums   []string
	Members       []string
	Locations     []string
}

func GetBandInfo(ArtistAPI string) ([]Band, error) {
	var bands []Band

	respArtist, err := http.Get(ArtistAPI)
	if err != nil {
		return nil, err
	}

	defer respArtist.Body.Close()

	err = json.NewDecoder(respArtist.Body).Decode(&bands)

	if err != nil {
		return nil, err
	}

	return bands, nil
}

func GetRelationsInfo(RelationsAPI string) (Relations, error) {
	var relations Relations

	respRelations, err := http.Get(RelationsAPI)
	if err != nil {
		return Relations{}, err
	}

	err = json.NewDecoder(respRelations.Body).Decode(&relations)

	if err != nil {
		return Relations{}, err
	}

	defer respRelations.Body.Close()

	return relations, nil
}

func GetLocationsInfo(LocationsAPI string) (Location, error) {
	var locations Location

	respLocations, err := http.Get(LocationsAPI)
	if err != nil {
		return Location{}, err
	}

	defer respLocations.Body.Close()

	err = json.NewDecoder(respLocations.Body).Decode(&locations)

	if err != nil {
		return Location{}, err
	}

	return locations, nil
}
