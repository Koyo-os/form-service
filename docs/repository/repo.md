package repository // import "github.com/Koyo-os/form-service/internal/repository"

Package repository provides data persistence functionality using GORM

TYPES

type Repository struct {
        // Has unexported fields.
}
    Repository handles database operations using GORM

func Init(db *gorm.DB, logger *logger.Logger) *Repository
    Init creates and returns a new Repository instance

func (repo *Repository) Create(payload any) error
    Create persists a new entity in the database Parameters:
      - payload: Any struct that maps to a database table

    Returns error if the creation fails

func (repo *Repository) DeleteForm(formID uuid.UUID) error
    DeleteForm removes a form from the database Parameters:
      - formID: UUID of the form to delete

    Returns error if the deletion fails

func (repo *Repository) DeleteQuestion(formID uuid.UUID, orderNumber uint) error
    DeleteQuestion removes a question from a form Parameters:
      - formID: UUID of the form containing the question
      - orderNumber: Position of the question in the form

    Returns error if the deletion fails

func (repo *Repository) Get(ID uuid.UUID) (*entity.Form, error)
    Get retrieves a form by its ID Parameters:
      - ID: UUID of the form to retrieve

    Returns:
      - *entity.Form: Retrieved form or nil if not found
      - error: Any error that occurred during retrieval

func (repo *Repository) Update(ID uuid.UUID, key string, value any) error
    Update modifies a single column of a form Parameters:
      - ID: UUID of the form to update
      - key: Column name to update
      - value: New value for the column

    Returns error if the update fails

func (repo *Repository) UpdateMany(ID uuid.UUID, value any) error
    UpdateMany updates multiple columns of a form simultaneously Parameters:
      - ID: UUID of the form to update
      - value: Struct containing the columns and values to update

    Returns error if the update fails

func (repo *Repository) UpdateQuestion(id uuid.UUID, key string, value any) error
    UpdateQuestion modifies a single column of a question Parameters:
      - id: UUID of the question to update
      - key: Column name to update
      - value: New value for the column

    Returns error if the update fails

func (repo *Repository) UpdateQuestionMany(id uuid.UUID, value any) error
    UpdateQuestionMany updates multiple columns of a question simultaneously
    Parameters:
      - id: UUID of the question to update
      - value: Struct containing the columns and values to update

    Returns error if the update fails