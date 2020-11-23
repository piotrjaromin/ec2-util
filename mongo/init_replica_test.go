package mongo

import (
	"errors"
	"testing"
	"time"

	mocks "github.com/piotrjaromin/ec2-util/mongo/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestFirstTimeInitShouldSkipWhenOldestHostHasMoreThan15Min(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	session := mocks.NewMockmongoSession(ctrl)

	hosts := map[string]time.Time{
		"host1": time.Now().Add(time.Duration(-5) * time.Minute),
		"host2": time.Now().Add(time.Duration(-25) * time.Minute),
		"host3": time.Now().Add(time.Duration(-12) * time.Minute),
	}

	res := firstTimeInit(session, "host1", hosts)

	assert.Nil(t, res)
}

func TestFirstTimeInitShouldSkipWhenOldestIsNotCurrent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	session := mocks.NewMockmongoSession(ctrl)

	hosts := map[string]time.Time{
		"host1": time.Now().Add(time.Duration(-5) * time.Minute),
		"host2": time.Now().Add(time.Duration(-6) * time.Minute),
		"host3": time.Now().Add(time.Duration(-9) * time.Minute),
	}

	res := firstTimeInit(session, "host1", hosts)

	assert.Nil(t, res)
}

func TestFirstTimeInitShouldIntReplicaSetWhenOldestIsCurrent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	session := mocks.NewMockmongoSession(ctrl)
	session.EXPECT().Run(gomock.Any(), gomock.Any()).Return(nil)

	hosts := map[string]time.Time{
		"host1": time.Now().Add(time.Duration(-5) * time.Minute),
		"host2": time.Now().Add(time.Duration(-6) * time.Minute),
		"host3": time.Now().Add(time.Duration(-9) * time.Minute),
	}

	res := firstTimeInit(session, "host3", hosts)

	assert.Nil(t, res)
}

func TestFirstTimeInitShouldPropagateMongoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	session := mocks.NewMockmongoSession(ctrl)

	err := errors.New("Some error from test")
	session.EXPECT().Run(gomock.Any(), gomock.Any()).Return(err)

	hosts := map[string]time.Time{
		"host1": time.Now().Add(time.Duration(-5) * time.Minute),
		"host2": time.Now().Add(time.Duration(-6) * time.Minute),
		"host3": time.Now().Add(time.Duration(-9) * time.Minute),
	}

	res := firstTimeInit(session, "host3", hosts)

	assert.Equal(t, err, res)
}
