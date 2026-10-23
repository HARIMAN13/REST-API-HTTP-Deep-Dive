package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"api-students/app/model"
)

var (
	ErrProjectLimit = errors.New("jumlah proyek sudah mencapai batas")
)

const MaxProjectsPerPortfolio = 20

type PortfolioRepository interface {
	EnsureIndexes(ctx context.Context) error
	FindByStudentID(ctx context.Context, studentID int) (model.Portfolio, error)
	Upsert(ctx context.Context, studentID int, req model.UpsertPortfolioRequest) (model.Portfolio, error)
	AddProject(ctx context.Context, studentID int, req model.CreateProjectRequest) (model.Portfolio, error)
	DeleteProject(ctx context.Context, studentID int, projectID string) error
	TopTechnologies(ctx context.Context, limit int) ([]model.TechnologyStat, error)
}

type portfolioDocument struct {
	ID           bson.ObjectID         `bson:"_id,omitempty"`
	StudentID    int                   `bson:"student_id"`
	Headline     string                `bson:"headline"`
	Bio          string                `bson:"bio"`
	Projects     []projectDocument     `bson:"projects"`
	Certificates []certificateDocument `bson:"certificates"`
	CreatedAt    time.Time             `bson:"created_at"`
	UpdatedAt    time.Time             `bson:"updated_at"`
}

type projectDocument struct {
	ID           bson.ObjectID `bson:"_id"`
	Title        string        `bson:"title"`
	Technologies []string      `bson:"technologies"`
	Link         string        `bson:"link"`
	Year         int           `bson:"year"`
}

type certificateDocument struct {
	ID     bson.ObjectID `bson:"_id"`
	Name   string        `bson:"name"`
	Issuer string        `bson:"issuer"`
	Date   time.Time     `bson:"date"`
}

func (d portfolioDocument) toModel() model.Portfolio {
	projects := make([]model.Project, 0, len(d.Projects))
	for _, p := range d.Projects {
		techs := p.Technologies
		if techs == nil {
			techs = []string{}
		}
		projects = append(projects, model.Project{
			ID:           p.ID.Hex(),
			Title:        p.Title,
			Technologies: techs,
			Link:         p.Link,
			Year:         p.Year,
		})
	}

	certs := make([]model.Certificate, 0, len(d.Certificates))
	for _, c := range d.Certificates {
		certs = append(certs, model.Certificate{
			ID:     c.ID.Hex(),
			Name:   c.Name,
			Issuer: c.Issuer,
			Date:   c.Date,
		})
	}

	return model.Portfolio{
		ID:           d.ID.Hex(),
		StudentID:    d.StudentID,
		Headline:     d.Headline,
		Bio:          d.Bio,
		Projects:     projects,
		Certificates: certs,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}

type portfolioMongoRepository struct {
	coll *mongo.Collection
}

func NewPortfolioRepository(db *mongo.Database) PortfolioRepository {
	return &portfolioMongoRepository{coll: db.Collection("student_portfolios")}
}

func (r *portfolioMongoRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "student_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("portfolio_student_unique"),
		},
		{
			Keys:    bson.D{{Key: "projects.technologies", Value: 1}},
			Options: options.Index().SetName("portfolio_project_technologies"),
		},
	})
	if err != nil {
		return fmt.Errorf("membuat index student_portfolios: %w", err)
	}
	return nil
}

func (r *portfolioMongoRepository) FindByStudentID(ctx context.Context, studentID int) (model.Portfolio, error) {
	var doc portfolioDocument
	err := r.coll.FindOne(ctx, bson.M{"student_id": studentID}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return model.Portfolio{}, ErrNotFound
		}
		return model.Portfolio{}, fmt.Errorf("mengambil portofolio: %w", err)
	}
	return doc.toModel(), nil
}

func (r *portfolioMongoRepository) Upsert(ctx context.Context, studentID int, req model.UpsertPortfolioRequest) (model.Portfolio, error) {
	t := now()
	projects := make([]projectDocument, 0, len(req.Projects))
	for _, p := range req.Projects {
		techs := p.Technologies
		if techs == nil {
			techs = []string{}
		}
		projects = append(projects, projectDocument{
			ID:           bson.NewObjectID(),
			Title:        p.Title,
			Technologies: techs,
			Link:         p.Link,
			Year:         p.Year,
		})
	}

	certs := make([]certificateDocument, 0, len(req.Certificates))
	for _, c := range req.Certificates {
		parsedDate, err := time.Parse("2006-01-02", c.Date)
		if err != nil {
			parsedDate = t
		}
		certs = append(certs, certificateDocument{
			ID:     bson.NewObjectID(),
			Name:   c.Name,
			Issuer: c.Issuer,
			Date:   parsedDate,
		})
	}

	filter := bson.M{"student_id": studentID}
	update := bson.M{
		"$set": bson.M{
			"headline":     req.Headline,
			"bio":          req.Bio,
			"projects":     projects,
			"certificates": certs,
			"updated_at":   t,
		},
		"$setOnInsert": bson.M{
			"_id":        bson.NewObjectID(),
			"student_id": studentID,
			"created_at": t,
		},
	}

	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	var doc portfolioDocument
	err := r.coll.FindOneAndUpdate(ctx, filter, update, opts).Decode(&doc)
	if err != nil {
		return model.Portfolio{}, fmt.Errorf("menyimpan portofolio: %w", err)
	}
	return doc.toModel(), nil
}

func (r *portfolioMongoRepository) AddProject(ctx context.Context, studentID int, req model.CreateProjectRequest) (model.Portfolio, error) {
	t := now()
	doc := projectDocument{
		ID:           bson.NewObjectID(),
		Title:        req.Title,
		Technologies: req.Technologies,
		Link:         req.Link,
		Year:         req.Year,
	}
	if doc.Technologies == nil {
		doc.Technologies = []string{}
	}

	limitField := "projects." + strconv.Itoa(MaxProjectsPerPortfolio-1)
	filter := bson.M{
		"student_id": studentID,
		limitField:   bson.M{"$exists": false},
	}
	update := bson.M{
		"$push": bson.M{"projects": doc},
		"$set":  bson.M{"updated_at": t},
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updated portfolioDocument
	err := r.coll.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updated)
	if err == nil {
		return updated.toModel(), nil
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return model.Portfolio{}, fmt.Errorf("menambah proyek: %w", err)
	}

	count, countErr := r.coll.CountDocuments(ctx, bson.M{"student_id": studentID})
	if countErr != nil {
		return model.Portfolio{}, fmt.Errorf("memeriksa portofolio: %w", countErr)
	}
	if count == 0 {
		return model.Portfolio{}, ErrNotFound
	}
	return model.Portfolio{}, ErrProjectLimit
}

func (r *portfolioMongoRepository) DeleteProject(ctx context.Context, studentID int, projectID string) error {
	oid, err := parseID(projectID)
	if err != nil {
		return err
	}

	filter := bson.M{"student_id": studentID, "projects._id": oid}
	update := bson.M{
		"$pull": bson.M{"projects": bson.M{"_id": oid}},
		"$set":  bson.M{"updated_at": now()},
	}

	res, err := r.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("menghapus proyek: %w", err)
	}
	if res.MatchedCount == 0 {
		count, countErr := r.coll.CountDocuments(ctx, bson.M{"student_id": studentID})
		if countErr != nil {
			return fmt.Errorf("memeriksa portofolio: %w", countErr)
		}
		if count == 0 {
			return ErrNotFound
		}
		return ErrNotFound
	}
	return nil
}

func (r *portfolioMongoRepository) TopTechnologies(ctx context.Context, limit int) ([]model.TechnologyStat, error) {
	if limit <= 0 {
		limit = 10
	}
	pipeline := mongo.Pipeline{
		{{Key: "$unwind", Value: "$projects"}},
		{{Key: "$unwind", Value: "$projects.technologies"}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$projects.technologies",
			"total": bson.M{"$sum": 1},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "total", Value: -1}, {Key: "_id", Value: 1}}}},
		{{Key: "$limit", Value: limit}},
	}

	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("agregasi teknologi: %w", err)
	}
	defer cursor.Close(ctx)

	var out []struct {
		Technology string `bson:"_id"`
		Total      int    `bson:"total"`
	}
	if err := cursor.All(ctx, &out); err != nil {
		return nil, fmt.Errorf("membaca hasil agregasi teknologi: %w", err)
	}

	results := make([]model.TechnologyStat, 0, len(out))
	for _, item := range out {
		results = append(results, model.TechnologyStat{
			Technology: item.Technology,
			Total:      item.Total,
		})
	}
	return results, nil
}
