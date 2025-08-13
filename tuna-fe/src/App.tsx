import React from "react";
import { Container, Typography, Box } from "@mui/material";
import DeployForm from "./DeployForm.tsx";
import DeploymentList from "./DeploymentList.tsx";

const App: React.FC = () => {
  return (
    <Container maxWidth="md">
      <Box my={4}>
        <Typography variant="h3" gutterBottom>Tuna Deployer</Typography>
        <DeployForm />
        <Typography variant="h5" gutterBottom style={{ marginTop: "2rem" }}>
          Deployments
        </Typography>
        <DeploymentList />
      </Box>
    </Container>
  );
};

export default App;
