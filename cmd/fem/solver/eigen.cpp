#include <Eigen/Sparse>
#ifdef __linux__
    #include <Eigen/PardisoSupport>
#else
    #include <Eigen/SparseCholesky>
#endif
#include <iostream>
#include "eigen.h"

using namespace Eigen;
using namespace std;

SparseMatrix<double> mat;
VectorXd vec;

void InitMatrix(int size, int max_non_zero)
{
    mat.resize(size, size);
    mat.setZero();
    vec.resize(size);
    vec.setZero();
    // Reserving memory
    mat.reserve(VectorXi::Constant(size, max_non_zero));
}

void SetMatrix(int row, int col, double value)
{
    mat.coeffRef(row, col) = value;
}

void AddMatrix(int row, int col, double value)
{
    mat.coeffRef(row, col) += value;
}

void SetVector(int i, double value)
{
    vec(i) = value;
}

void AddVector(int i, double value)
{
    vec(i) += value;
}

double GetMatrix(int row, int col)
{
    return mat.coeffRef(row, col);
}

double GetVector(int i)
{
    return vec(i);
}

void SetBoundaryCondition(int index, double value)
{
    for (Eigen::SparseMatrix<double>::InnerIterator i(mat, index); i; ++i)
    {
        if (i.row() not_eq i.col())
        {
            mat.coeffRef(i.row(), i.col()) = value;
            mat.coeffRef(i.col(), i.row()) = value;
        }
    }
    vec(index) = value * mat.coeffRef(index, index);
}

int SolveEigen(double *res)
{
#ifdef __linux__
    PardisoLLT<SparseMatrix<double>> solver;
#else
    SimplicialLLT<SparseMatrix<double>> solver;
#endif
    solver.compute(mat);
    if (solver.info() != Success)
        return 1;
    vec = solver.solve(vec);
    for (auto i = 0u; i < mat.rows(); i++)
        res[i] = vec[i];
    mat.resize(0, 0);
    vec.resize(0);
    return 0;
}