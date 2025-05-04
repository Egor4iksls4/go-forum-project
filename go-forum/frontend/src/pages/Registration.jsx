import { useState, useEffect } from "react";
import { useNavigate, Link } from "react-router-dom";
import api from "../api/api";
import "./Registration.css";

export default function Registration() {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');
    const navigate = useNavigate();

    useEffect(() => {
        if (localStorage.getItem("isAuth") === "true") {
            navigate("/profile");
        }
    }, [navigate]);

    const handleSubmit = async (e) => {
        e.preventDefault();
        setError("");

        try {
            const response = await api.post("/register", {
                username,
                password,
            });

            if (response.status === 200) {
                localStorage.setItem("isAuth", "true");
                navigate("profile");
            }
        } catch (err) {
            if (err.response) {
                if (err.response.status === 409) {
                    setError("Пользователь с таким именем уже существует");
                } else {
                    setError("Ошибка регистрации. Проверьте данные!");
                }
            } else {
                setError("Сервер не отвечает.");
            }
        }
    };

    return (
        <div className="box">
            <h1>Регистрация</h1>
            <form onSubmit={handleSubmit} className="form-style">
                <div>
                    <label>Имя пользователя: </label>
                    <input 
                        type="text"
                        value={username}
                        onChange={(e) => setUsername(e.target.value)}
                        className="input-style"
                        required
                    />
                </div>
                <div>
                    <label>Пароль: </label>
                    <input 
                        type="password"
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        className="input-style"
                        required
                    />
                </div>
                {error && <p className="error">{error}</p>}
                <button type="submit" className="button-style">Зарегистрироваться</button>
                <br></br>
                <br></br>
                <br></br>
                <div>
                    <Link to="/login">Вход</Link>
                </div>
            </form>
        </div>
    );
}