package leaderBoardModel

type UserLeaderBoard struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Points   string `json:"points"`
}
