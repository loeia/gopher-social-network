package avatar

import (
	"fmt"

	dicebear "github.com/dicebear/dicebear-go/v10"
	"github.com/dicebear/styles/v10"
)

var style *dicebear.Style

func Init() error {
	var err error
	style, err = dicebear.NewStyle([]byte(styles.Thumbs))
	return err
}

func Generate(userId int64) ([]byte, error) {
	seed := fmt.Sprintf("user-%d", userId)

	avatar, err := dicebear.NewAvatar(style, map[string]any{
		"seed":         seed,
		"size":         128,
		"borderRadius": 50,
	})
	if err != nil {
		return nil, err
	}

	return []byte(avatar.SVG()), nil
}
