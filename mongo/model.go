package mongo

type Config struct {
	Name    string         `bson:"_id"`
	Version int            `bson:"version"`
	Members []ConfigMember `bson:"members"`
}

type ConfigMember struct {
	ID      int    `bson:"_id"`
	Address string `bson:"host"`
}

type ReplSetStatus struct {
	Members []Member `bson:"members"`
	Set     string   `bson:"set"`
}

type Member struct {
	ID       int    `bson:"_id"`
	Name     string `bson:"name"`
	Self     bool   `bson:"self"`
	State    int    `bson:"state"`
	StateStr string `bson:"stateStr"`
	Health   int    `bson:"health"`
}

const healthyState = 1
