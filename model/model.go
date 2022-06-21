package model

type MakeFriends struct {
	Source int `json:"source_id"`
	Target int `json:"target_id"`
}

type FbStore struct {
}

type User struct {
	Id      int      `firebase:"Id"`
	Name    string   `json:"Name" firebase:"Name"`
	Age     int      `json:"Age" firebase:"Age"`
	Friends []string `json:"Friends" firebase:"Friends"`
}

type DeleteUser struct {
	Target int `json:"target_id"`
}

type AgeUpdate struct {
	Age int `json:"new_age"`
}
