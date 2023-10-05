package main

type Artist struct {
	ID           int
	Name         string
	CreationDate int
	FirstAlbum   string
	Image        string
	Members      []string
}

type Relation struct {
	Index []artistRelation
}

type Geodata struct {
	Index []struct {
		CountryCoords map[string][]string
	} `json:"index"`
}


type artistRelation struct {
	ID             int
	DatesLocations map[string]interface{}
}

type artistData struct {
	Artist         Artist
	DatesLocations map[string]interface{}
}
