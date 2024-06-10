package model

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/orisano/wyhash"

	"sealchat/utils"
)

var _tokenSecret = "token-secret"
var _tokenSeed uint64 = 0xCAFEBABE

var ErrInvalidToken = errors.New("invalid token")
var ErrTokenExpired = errors.New("token expired")

func SetTokenSecret(secret string, seed uint64) {
	if secret != "" {
		_tokenSecret = secret
	}
	if seed != 0 {
		_tokenSeed = seed
	}
}

func TokenSign(tokenBase string, expireAt time.Time) string {
	expStr := strconv.FormatInt(expireAt.Unix()/60, 35)
	mainStr := fmt.Sprintf("%s-%s", tokenBase, expStr)

	hash := wyhash.Sum64(_tokenSeed, []byte(mainStr+"-"+_tokenSecret))
	return fmt.Sprintf("%s-%x", mainStr, hash>>32)
}

func TokenGenerate(expireAt time.Time) string {
	return TokenSign(utils.NewID(), expireAt)
}

type TokenCheckResult struct {
	HashValid    bool
	TimeValid    bool
	ExpireOffset int64
	Token        string
}

func TokenCheck(x string) TokenCheckResult {
	var ret TokenCheckResult

	parts := strings.Split(x, "-")
	if len(parts) != 3 {
		return ret
	}
	ret.Token = parts[0]

	// check valid
	data := []byte(fmt.Sprintf("%s-%s-%s", parts[0], parts[1], _tokenSecret))
	hash := wyhash.Sum64(_tokenSeed, data)

	if fmt.Sprintf("%x", hash>>32) != parts[2] {
		return ret
	}

	ret.HashValid = true

	// check expire
	now := time.Now().Unix() / 60
	tokenExpire, err := strconv.ParseInt(parts[1], 35, 64)
	if err != nil {
		return ret
	}

	offset := tokenExpire - now
	ret.ExpireOffset = offset

	if offset < 0 {
		return ret
	}

	ret.TimeValid = true
	return ret
}
