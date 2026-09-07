package repository

import (
	"context"
	"api-students/app/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PrestasiRepository struct {
	p *pgxpool.Pool
}

func NewPrestasiRepository(pool *pgxpool.Pool) *PrestasiRepository {
	return &PrestasiRepository{p: pool}
}

func (r *PrestasiRepository) ListByStudentID(ctx context.Context, sID int) ([]model.Prestasi, error) {
	rows, err := r.p.Query(ctx, "SELECT id, nama_prestasi, student_id, juara FROM prestasi WHERE student_id = $1", sID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []model.Prestasi
	for rows.Next() {
		var p model.Prestasi
		if err := rows.Scan(&p.ID, &p.NamaPrestasi, &p.StudentID, &p.Juara); err != nil {
			return nil, err
		}
		data = append(data, p)
	}
	return data, nil
}
