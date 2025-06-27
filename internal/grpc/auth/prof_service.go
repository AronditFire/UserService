package profgrpc

type UserProfile interface {
}

type ProfileAPI struct {
	uProfile UserProfile
}

func NewProfileAPI(userProfile UserProfile) *ProfileAPI {
	return &ProfileAPI{uProfile: userProfile}
}
