import { useState, useEffect } from "react";
import { useNavigate, Link } from "react-router-dom";
import "./Login.css";
import api from "../api/api";

export default function Login(){
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');
    const navigate = useNavigate();

    useEffect(() => {
        if (localStorage.getItem('isAuth') === 'true') {
          navigate('/profile');
        }
      }, [navigate]);

    const handleSubmit = async (e) => {
        e.preventDefault();

        // try {
        //     const res = await api.post("/auth/login", {
        //         username,
        //         password
        //     });

        //     if (res.data.seccess) {
        //         localStorage.setItem('isAuth', 'true');
        //         navigate('/profile');
        //     } else {
        //         setError('Неверное имя пользователя или пароль!');
        //     }
        // } catch (err) {
        //     console.error(err);
        //     setError('Ошибка сервера при попытке входа!');
        // }
        if (username === 'admin' && password === '12345'){
            localStorage.setItem('isAuth', 'true');
            navigate('/profile');
        } else {
            setError('Имя пользователя или пароль введены неверно!');
        }
    };

    return (
        <div className="box">
            <h1>Вход</h1>
            <form onSubmit={handleSubmit} className="form-style">
                <div>
                    <label>Имя пользователя: </label>
                    <input
                        type="text"
                        value={username}
                        onChange={(e) => setUsername(e.target.value)}
                        className="input-style"
                    />
                </div>
                <div>
                    <label>Пароль: </label>
                    <input 
                        type="password"
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        className="input-style"
                    />
                </div>
                {error && <p className="error">{error}</p>}
                <button type="submit" className="button-style">Войти</button>
                <br></br>
                <br></br>
                <br></br>
                <div>
                    <Link to="/register">Регистрация</Link>
                </div>
            </form>
        </div>
    );
}