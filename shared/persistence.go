package shared

type DataProvider interface {
	GetId() string
	GetData() Complex
}

type Repository interface {
	Create(provider DataProvider) error

	Get(id, version string) (DataProvider, error)

	GetByUserName(username string) (DataProvider, error)

	GetAll() ([]Complex, error)

	Count(query string) (int, error)

	Update(id, version string, provider DataProvider) error

	Delete(id, version string) error

	Search(payload SearchRequest) (*ListResponse, error)
}

// An simple in memory database fit for test use and read only production use
// this implementation:
// - only implements Create, Get, GetAll, Update, Delete
// - is not thread safe
// - ignores the version argument
type mapRepository struct {
	data map[string]DataProvider
}

func (r *mapRepository) Create(provider DataProvider) error {
	r.data[provider.GetId()] = provider
	return nil
}

func (r *mapRepository) Get(id, version string) (DataProvider, error) {
	if dp, ok := r.data[id]; !ok {
		return nil, Error.ResourceNotFound(id, version)
	} else {
		return dp, nil
	}
}

func (r *mapRepository) GetByUsername(username string) (DataProvider, error) {
	if dp, ok := r.data[username]; !ok {
		return nil, Error.Text("Resource not found searching by username")
		//TODO: Error.ResourceNotFound related to username, and not only id and version
	} else {
		return dp, nil
	}
}

func (r *mapRepository) GetAll() ([]Complex, error) {
	all := make([]Complex, 0)
	for _, v := range r.data {
		all = append(all, v.GetData())
	}
	return all, nil
}

func (r *mapRepository) Count(query string) (int, error) {
	return 0, Error.Text("not implemented")
}

func (r *mapRepository) Update(id, version string, provider DataProvider) error {
	if _, ok := r.data[id]; !ok {
		return Error.ResourceNotFound(id, version)
	} else {
		r.data[id] = provider
		return nil
	}
}

func (r *mapRepository) Delete(id, version string) error {
	if _, ok := r.data[id]; !ok {
		return Error.ResourceNotFound(id, version)
	} else {
		delete(r.data, id)
		return nil
	}
}

func (r *mapRepository) Search(payload SearchRequest) (*ListResponse, error) {
	return nil, Error.Text("not implemented")
}

func NewMapRepository(initialData map[string]DataProvider) Repository {
	if len(initialData) == 0 {
		return &mapRepository{data: make(map[string]DataProvider, 0)}
	} else {
		return &mapRepository{data: initialData}
	}
}

// simple factory method to return a repository query method
// to easily put together a query only repository by composing
// several repositories, useful when implementing root query functions
func CompositeSearchFunc(
	repositories ...Repository,
) func(payload SearchRequest) (*ListResponse, error) {
	return func(payload SearchRequest) (*ListResponse, error) {
		// prepare plans
		skipQuota, limitQuota := 0, 0
		grandListResponse := &ListResponse{
			Schemas:      []string{ListResponseUrn},
			StartIndex:   payload.StartIndex,
			ItemsPerPage: payload.Count,
			TotalResults: 0,
			Resources:    make([]DataProvider, 0),
		}

		// set parameter defaults
		if payload.StartIndex < 1 {
			skipQuota = 0
		} else {
			skipQuota = payload.StartIndex - 1
		}

		if payload.Count > 0 {
			limitQuota = payload.Count
		} else {
			limitQuota = 0
		}

		// generate plan structures
		plans := make([]*queryExecutionPlan, 0, len(repositories))
		for _, n := range repositories {
			plans = append(plans, &queryExecutionPlan{repo: n})
		}
		for _, plan := range plans {
			count, err := plan.repo.Count(payload.Filter)
			if err != nil {
				plan.skipAll = true
			} else {
				plan.resultCount = count
			}
			grandListResponse.TotalResults += count
		}

		// devise plans
		for _, plan := range plans {
			if plan.skipAll {
				continue
			}

			if limitQuota <= 0 {
				plan.skipAll = true
				continue
			}

			if plan.resultCount <= 0 {
				plan.skipAll = true
				continue
			}

			if skipQuota <= 0 {
				// take a look at limit
				plan.skipAll = false
				plan.skip = 0
				if limitQuota < plan.resultCount {
					plan.limit = limitQuota
				} else {
					plan.limit = plan.resultCount
				}
				limitQuota -= plan.limit
			} else {
				// continue to skip
				if plan.resultCount > skipQuota {
					// partial results needed
					plan.skipAll = false
					plan.skip = skipQuota
					if limitQuota < plan.resultCount-plan.skip {
						plan.limit = limitQuota
					} else {
						plan.limit = plan.resultCount - plan.skip
					}
					limitQuota -= plan.limit
				} else {
					plan.skipAll = true
				}
				skipQuota -= plan.resultCount
			}
		}

		// execution plan
		for _, plan := range plans {
			if plan.skipAll || plan.limit == 0 {
				continue
			}

			sr := SearchRequest{
				Filter:             payload.Filter,
				StartIndex:         plan.skip,
				Count:              plan.limit,
				SortBy:             payload.SortBy,
				SortOrder:          payload.SortOrder,
				Attributes:         payload.Attributes,
				ExcludedAttributes: payload.ExcludedAttributes,
				Schemas:            payload.Schemas,
			}
			if sr.StartIndex < 1 {
				sr.StartIndex = 1
			}
			if sr.Count < 0 {
				sr.Count = 0
			}

			listResp, err := plan.repo.Search(sr)
			if err != nil {
				return nil, err
			}
			grandListResponse.Resources = append(grandListResponse.Resources, listResp.Resources...)
		}

		return grandListResponse, nil
	}
}

type queryExecutionPlan struct {
	repo        Repository
	resultCount int
	skipAll     bool
	skip        int
	limit       int
}
