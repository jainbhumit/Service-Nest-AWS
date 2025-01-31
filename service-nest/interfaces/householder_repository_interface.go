package interfaces

import "service-nest/model"

type HouseholderRepository interface {
	SaveHouseholder(householder *model.Householder) error
}
