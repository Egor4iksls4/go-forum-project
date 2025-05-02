import { useNavigate } from "react-router-dom";
import "./Profile.css";

export default function Profile(){
    const navigate = useNavigate();

    const handleLogout = () => {
        localStorage.setItem('isAuth', 'false');
        navigate('/login');
    };

    return (
        <div className="box">
            <h1>Профиль</h1>
            <p>Это страница вашего профиля.</p>
            <button onClick={handleLogout} className="out-style">Выйти из профиля</button>
        </div>
    );
}