#ifndef EIGEN_H
#define EIGEN_H

// __cplusplus gets defined when a C++ compiler processes the file.
// extern "C" is needed so the C++ compiler exports the symbols w/out name issues.
#ifdef __cplusplus
extern "C" {
#endif

void InitMatrix(int, int);
void SetBoundaryCondition(int, double);
void SetMatrix(int, int, double);
void AddMatrix(int, int, double);
void SetVector(int, double);
void AddVector(int, double);
double GetMatrix(int, int);
double GetVector(int);

int SolveEigen(double*);

#ifdef __cplusplus
}
#endif

#endif // EIGEN_H