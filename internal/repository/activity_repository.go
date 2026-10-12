package repository

import (
	"github.com/nim-sam/gitport/internal/model"
	orm "github.com/nim-sam/gitport/pkg/mini_orm"
)

// Activity Repository
var ActivityRepo = orm.Repository[model.Activity]{}
