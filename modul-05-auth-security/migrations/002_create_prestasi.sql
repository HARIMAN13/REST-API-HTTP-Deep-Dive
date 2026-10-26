CREATE TABLE prestasi (
    id SERIAL PRIMARY KEY,
    nama_prestasi VARCHAR(255) NOT NULL,
    student_id INT NOT NULL,
    juara VARCHAR(50) NOT NULL
);
