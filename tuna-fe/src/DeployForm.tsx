import React, { useState } from "react";
import { Button, TextField, Select, MenuItem, FormControl, InputLabel, Box } from "@mui/material";
import axios from "axios";
import type { DeployRequest } from "./types";
import { API_BASE_URL } from "./consts";

const DeployForm: React.FC = () => {
  const [formData, setFormData] = useState<DeployRequest>({
    name: "",
    mode: "build",
    repo: "",
    branch: "",
    template: "",
    image: "",
    env: {},
    ingress: "",
  });

  const [envInput, setEnvInput] = useState<string>("{}");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const env = JSON.parse(envInput);
      const response = await axios.post(API_BASE_URL + "/deploy", { ...formData, env });
      alert("Deployed: " + response.data.container_id);
      window.location.reload(); // Refresh to update deployment list
    } catch (error) {
      const e = error as any;
      alert("Error: " + (e?.response?.data || e?.message));
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | { name?: string; value: unknown }>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name!]: value }));
  };

  return (
    <Box component="form" onSubmit={handleSubmit} sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
      <TextField name="name" label="Deployment Name" required onChange={handleChange} />
      <FormControl>
        <InputLabel>Mode</InputLabel>
        <Select name="mode" value={formData.mode} onChange={handleChange}>
          <MenuItem value="build">Build</MenuItem>
          <MenuItem value="image">Image</MenuItem>
        </Select>
      </FormControl>
      {formData.mode === "build" && (
        <>
          <TextField name="repo" label="Git Repository URL" onChange={handleChange} />
          <TextField name="branch" label="Branch" onChange={handleChange} />
          <FormControl>
            <InputLabel>Template</InputLabel>
            <Select name="template" value={formData.template} onChange={handleChange}>
              <MenuItem value="">None (use Dockerfile)</MenuItem>
              <MenuItem value="static">Static</MenuItem>
              <MenuItem value="nextjs">Next.js</MenuItem>
              <MenuItem value="gradle">Gradle/Java</MenuItem>
              <MenuItem value="rubyonrails">Ruby on Rails</MenuItem>
            </Select>
          </FormControl>
        </>
      )}
      {formData.mode === "image" && (
        <TextField name="image" label="Docker Image (e.g., nginx:latest)" onChange={handleChange} />
      )}
      <TextField
        name="env"
        label="Environment Variables (JSON)"
        multiline
        rows={4}
        value={envInput}
        onChange={(e) => setEnvInput(e.target.value)}
      />
      <TextField name="ingress" label="Ingress Host (e.g., app.example.com)" required onChange={handleChange} />
      <Button type="submit" variant="contained" color="primary">Deploy</Button>
    </Box>
  );
};

export default DeployForm;
