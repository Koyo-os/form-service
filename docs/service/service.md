TYPES

type Casher interface {
        AddToCash(ctx context.Context, key string, payload any) error // payload must be pointer
        GetCashFor(ctx context.Context, key string) ([]byte, error)
        RemoveFromCash(ctx context.Context, key string) error
}

type Publisher interface {
        Publish(any, string) error
}

type Repository interface {
        Create(any) error
        Update(uuid.UUID, string, any) error
        UpdateMany(uuid.UUID, any) error
        Get(uuid.UUID) (*entity.Form, error)
        DeleteForm(uuid.UUID) error
        DeleteQuestion(uuid.UUID, uint) error
}

type Service struct {
        // Has unexported fields.
}
    Service provides business logic for form management operations. It
    coordinates between repository, cache, and event publishing systems.

func Init(casher Casher, repo Repository, publisher Publisher, timeout time.Duration) *Service
    Init initializes and returns a new Service instance with dependencies.
    It sets up a context with a default 10-second timeout for all service
    operations. Parameters:
      - casher: Cache handler implementation
      - repo: Repository implementation for data access
      - publisher: Event publisher implementation

    Returns:
      - *Service: Initialized service instance

func (s *Service) CreateForm(form *entity.Form) error
    CreateForm creates a new form in the system. It performs the following
    operations:
     1. Persists the form in the repository
     2. Publishes a "form.created" event
     3. Caches the form data (with retry logic)

    Parameters:
      - form: Pointer to the Form entity to create

    Returns:
      - error: Any error that occurs during the operation

func (s *Service) CreateQuestion(question *entity.Question) error
    CreateQuestion adds a new question to an existing form. It performs the
    following operations:
     1. Persists the question in the repository
     2. Retrieves the updated form
     3. Publishes a "form.updated" event (with retry logic)
     4. Updates the form in cache (with retry logic)

    Parameters:
      - question: Pointer to the Question entity to create

    Returns:
      - error: Any error that occurs during the operation

func (s *Service) DeleteForm(formId string) error
    DeleteForm removes a form from the system. It performs the following
    operations:
     1. Deletes the form from the repository
     2. Removes the form from cache (with retry logic)
     3. Publishes a "form.deleted" event (with retry logic)

    Parameters:
      - formId: String UUID of the form to delete

    Returns:
      - error: Any error that occurs during the operation

func (s *Service) DeleteQuestion(formId string, orderNumber uint) error
    DeleteQuestion removes a question from a form. It performs the following
    operations:
     1. Deletes the question from the repository
     2. Retrieves the updated form
     3. Updates the form in cache (with retry logic)
     4. Publishes a "form.updated" event

    Parameters:
      - formId: String UUID of the form containing the question
      - orderNumber: The order number of the question to delete

    Returns:
      - error: Any error that occurs during the operation

func (s *Service) Update(formID uuid.UUID, values any) error
    Update modifies multiple fields of a form at once. It performs the following
    operations:
     1. Updates fields in the repository
     2. Updates the cache with new values (with retry logic)
     3. Publishes a "form.updated" event (with retry logic)

    Parameters:
      - formID: UUID of the form to update
      - values: Interface containing the new field values

    Returns:
      - error: Any error that occurs during the operation

func (s *Service) UpdateDescription(formId string, desc string) error
    UpdateDescription changes the description of a form. It performs the
    following operations:
     1. Updates the description in the repository
     2. Retrieves the updated form
     3. Updates the form in cache (with retry logic)
     4. Publishes a "form.updated" event (with retry logic)

    Parameters:
      - formId: String UUID of the form to update
      - desc: New description text

    Returns:
      - error: Any error that occurs during the operation

func (s *Service) UpdateStatus(form_id string, closed bool) error
    UpdateStatus changes the closed/open status of a form. It performs the
    following operations:
     1. Updates the status in the repository
     2. Retrieves the updated form
     3. Updates the form in cache (with retry logic)
     4. Publishes a "form.created" event (note: potentially should be
        "form.updated")

    Parameters:
      - form_id: String UUID of the form to update
      - closed: Boolean indicating new status (true = closed)

    Returns:
      - error: Any error that occurs during the operation