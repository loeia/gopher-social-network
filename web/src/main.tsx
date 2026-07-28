import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
import { createBrowserRouter, RouterProvider } from "react-router-dom";
import { ConfirmationPage } from "./confirmationPage";
import App from "./App";

const router = createBrowserRouter([
    {
        path: "/",
        element: <App />,
    },
    {
        path: "/confirm/:token",
        element: <ConfirmationPage />,
    },
]);

createRoot(document.getElementById("root")!).render(
    <StrictMode>
        <RouterProvider router={router} />
    </StrictMode>,
);
