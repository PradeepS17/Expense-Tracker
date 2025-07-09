import React, { useEffect, useState } from "react";
import axios from "axios";
import "./App.css";

const API_URL = "http://localhost:8080";

export default function App() {
  const [expenses, setExpenses] = useState([]);
  const [form, setForm] = useState({
    title: "",
    amount: "",
    category: "",
    description: "",
  });
  const [editId, setEditId] = useState(null);
  const [categories, setCategories] = useState([]);
  const [filterCategory, setFilterCategory] = useState("");

  const fetchExpenses = async (category = "") => {
    try {
      const res = await axios.get(
        `${API_URL}/expense${category ? `?category=${encodeURIComponent(category)}` : ""}`
      );
      setExpenses(Array.isArray(res.data) ? res.data : []);
    } catch (e) {
      setExpenses([]);
      console.error("Fetch error:", e);
    }
  };

  const fetchCategories = async () => {
    try {
      const res = await axios.get(`${API_URL}/categories`);
      setCategories(Array.isArray(res.data) ? res.data : []);
    } catch (e) {
      setCategories([]);
    }
  };

  useEffect(() => {
    fetchExpenses(filterCategory);
    fetchCategories();
  }, [filterCategory]);

  const handleChange = (e) => {
    setForm({ ...form, [e.target.name]: e.target.value });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!form.title || !form.amount) return;
    try {
      if (editId) {
        await axios.put(`${API_URL}/expense/${editId}`, {
          ...form,
          amount: parseFloat(form.amount),
        });
        setEditId(null);
      } else {
        await axios.post(`${API_URL}/expense`, {
          ...form,
          amount: parseFloat(form.amount),
        });
      }
      setForm({ title: "", amount: "", category: "", description: "" });
      fetchExpenses(filterCategory);
      fetchCategories();
    } catch (e) {
      alert("Error: " + (e.response?.data?.error || e.message));
    }
  };

  const handleEdit = (exp) => {
    setEditId(exp.id || exp._id);
    setForm({
      title: exp.title,
      amount: exp.amount,
      category: exp.category,
      description: exp.description,
    });
  };

  const handleDelete = async (id) => {
    if (!window.confirm("Delete this expense?")) return;
    try {
      await axios.delete(`${API_URL}/expense/${id}`);
      fetchExpenses(filterCategory);
      fetchCategories();
    } catch (e) {
      alert("Delete failed");
    }
  };

  return (
    <div className="container">
      <h1 className="heading">💰 Expense Tracker</h1>

      <form className="expense-form" onSubmit={handleSubmit}>
        <input name="title" placeholder="Title" value={form.title} onChange={handleChange} required />
        <input name="amount" type="number" step="1" placeholder="Amount" value={form.amount} onChange={handleChange} required />
        <input list="category-list" name="category" placeholder="Category" value={form.category} onChange={handleChange} required />
        <datalist id="category-list">
          {categories.map((cat) => (
            <option value={cat} key={cat} />
          ))}
        </datalist>
        <textarea name="description" placeholder="Description" value={form.description} onChange={handleChange} rows={2} />
        <div className="form-buttons">
          <button type="submit" className="btn submit-btn">
            {editId ? "Update" : "Add"}
          </button>
          {editId && (
            <button type="button" className="btn cancel-btn" onClick={() => {
              setEditId(null);
              setForm({ title: "", amount: "", category: "", description: "" });
            }}>
              Cancel
            </button>
          )}
        </div>
      </form>

      <div className="filter-bar">
        <label>Filter by Category:</label>
        <select value={filterCategory} onChange={(e) => setFilterCategory(e.target.value)}>
          <option value="">All</option>
          {categories.map((cat) => (
            <option value={cat} key={cat}>{cat}</option>
          ))}
        </select>
      </div>

      <ul className="expense-list">
        {expenses.map((exp) => (
          <li key={exp.id || exp._id} className="expense-item">
            <div className="expense-header">
              <strong>{exp.title}</strong>
              <span className="amount">₹{exp.amount}</span>
            </div>
            <div className="expense-details">
              <div><strong>Category:</strong> {exp.category || "N/A"}</div>
              <div><strong>Date:</strong> {exp.date ? new Date(exp.date).toLocaleString() : "N/A"}</div>
              <div><strong>Description:</strong> {exp.description}</div>
            </div>
            <div className="expense-actions">
              <button onClick={() => handleEdit(exp)} className="btn edit-btn">Edit</button>
              <button onClick={() => handleDelete(exp.id || exp._id)} className="btn delete-btn">Delete</button>
            </div>
          </li>
        ))}
      </ul>
    </div>
  );
}
