import { Link } from "react-router-dom";
import "./Navigation.css";

export default function Navigation(){
    const isAuth = localStorage.getItem("isAuth") === "true";
    return (
        <header className="header-style">
            <nav className="navigation">
                <div className="main-links-box">
                    <Link to="/" className="nav-link">Главная</Link>
                    <Link to="/news" className="nav-link">Темы</Link>
                </div>
                <div>
                    {isAuth ? (
                        <Link to="/profile" className="nav-link">Профиль</Link>
                    ) : (
                        <Link to="/login" className="nav-link">Войти</Link>
                    )}
                </div>
            </nav>
        </header>
    );
}